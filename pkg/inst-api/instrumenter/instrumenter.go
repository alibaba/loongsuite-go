// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instrumenter

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Instrumenter[REQUEST any, RESPONSE any] interface {
	ShouldStart(parentContext context.Context, request REQUEST) bool
	StartAndEnd(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time)
	StartAndEndWithOptions(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time, startOptions []trace.SpanStartOption, endOptions []trace.SpanEndOption)
	Start(parentContext context.Context, request REQUEST, options ...trace.SpanStartOption) context.Context
	End(ctx context.Context, request REQUEST, response RESPONSE, err error, options ...trace.SpanEndOption)
}

type InternalInstrumenter[REQUEST any, RESPONSE any] struct {
	enabler              InstrumentEnabler
	spanNameExtractor    SpanNameExtractor[REQUEST]
	spanKindExtractor    SpanKindExtractor[REQUEST]
	spanStatusExtractor  SpanStatusExtractor[REQUEST, RESPONSE]
	attributesExtractors []AttributesExtractor[REQUEST, RESPONSE]
	operationListeners   []OperationListener
	contextCustomizers   []ContextCustomizer[REQUEST]
	spanSuppressor       SpanSuppressor
	tracer               trace.Tracer
	instVersion          string
}

type PropagatingToDownstreamInstrumenter[REQUEST any, RESPONSE any] struct {
	carrierGetter func(REQUEST) propagation.TextMapCarrier
	prop          propagation.TextMapPropagator
	base          InternalInstrumenter[REQUEST, RESPONSE]
}

type PropagatingFromUpstreamInstrumenter[REQUEST any, RESPONSE any] struct {
	carrierGetter func(REQUEST) propagation.TextMapCarrier
	prop          propagation.TextMapPropagator
	base          InternalInstrumenter[REQUEST, RESPONSE]
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) ShouldStart(parentContext context.Context, request REQUEST) bool {
	spanKind := i.spanKindExtractor.Extract(request)
	suppressed := i.spanSuppressor.ShouldSuppress(parentContext, spanKind)
	// TODO: record suppressed span
	return !suppressed
}

var cachePool = &sync.Pool{
	New: func() interface{} {
		return make([]attribute.KeyValue, 0, 25)
	},
}

func GetCachedAttrs() []attribute.KeyValue {
	return cachePool.Get().([]attribute.KeyValue)
}

func PutCachedAttrs(attrs []attribute.KeyValue) {
	attrs = attrs[:0]
	cachePool.Put(attrs)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) StartAndEnd(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time) {
	ctx := i.doStart(parentContext, request, startTime)
	i.doEnd(ctx, request, response, err, endTime)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) StartAndEndWithOptions(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time, startOptions []trace.SpanStartOption, endOptions []trace.SpanEndOption) {
	ctx := i.doStart(parentContext, request, startTime, startOptions...)
	i.doEnd(ctx, request, response, err, endTime, endOptions...)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) Start(parentContext context.Context, request REQUEST, options ...trace.SpanStartOption) context.Context {
	return i.doStart(parentContext, request, time.Now(), options...)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) doStart(parentContext context.Context, request REQUEST, timestamp time.Time, options ...trace.SpanStartOption) context.Context {
	if i.enabler != nil && !i.enabler.Enable() {
		return parentContext
	}
	for _, listener := range i.operationListeners {
		parentContext = listener.OnBeforeStart(parentContext, timestamp)
	}
	// extract span name
	spanName := i.spanNameExtractor.Extract(request)
	spanKind := i.spanKindExtractor.Extract(request)
	options = append(options, trace.WithSpanKind(spanKind), trace.WithTimestamp(timestamp))
	newCtx, span := i.tracer.Start(parentContext, spanName, options...)
	attrs := make([]attribute.KeyValue, 0, 20)
	// extract span attrs
	for _, extractor := range i.attributesExtractors {
		attrs, newCtx = extractor.OnStart(attrs, newCtx, request)
	}
	// execute context customizer hook
	for _, customizer := range i.contextCustomizers {
		newCtx = customizer.OnStart(newCtx, request, attrs)
	}
	for _, listener := range i.operationListeners {
		newCtx = listener.OnBeforeEnd(newCtx, attrs, timestamp)
	}
	span.SetAttributes(attrs...)
	return i.spanSuppressor.StoreInContext(newCtx, spanKind, span)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) End(ctx context.Context, request REQUEST, response RESPONSE, err error, options ...trace.SpanEndOption) {
	i.doEnd(ctx, request, response, err, time.Now(), options...)
}

func (i *InternalInstrumenter[REQUEST, RESPONSE]) doEnd(ctx context.Context, request REQUEST, response RESPONSE, err error, timestamp time.Time, options ...trace.SpanEndOption) {
	if i.enabler != nil && !i.enabler.Enable() {
		return
	}
	for _, listener := range i.operationListeners {
		listener.OnAfterStart(ctx, timestamp)
	}
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	attrs := GetCachedAttrs()
	defer PutCachedAttrs(attrs)
	// extract span attributes
	for _, extractor := range i.attributesExtractors {
		attrs, ctx = extractor.OnEnd(attrs, ctx, request, response, err)
	}
	i.spanStatusExtractor.Extract(span, request, response, err)
	span.SetAttributes(attrs...)
	options = append(options, trace.WithTimestamp(timestamp))
	span.End(options...)
	for _, listener := range i.operationListeners {
		listener.OnAfterEnd(ctx, attrs, timestamp)
	}
}

func (p *PropagatingToDownstreamInstrumenter[REQUEST, RESPONSE]) ShouldStart(parentContext context.Context, request REQUEST) bool {
	return p.base.ShouldStart(parentContext, request)
}
func (p *PropagatingToDownstreamInstrumenter[REQUEST, RESPONSE]) StartAndEnd(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time) {
	newCtx := p.base.doStart(parentContext, request, startTime)
	if p.carrierGetter != nil {
		if p.prop != nil {
			p.prop.Inject(newCtx, p.carrierGetter(request))
		} else {
			otel.GetTextMapPropagator().Inject(newCtx, p.carrierGetter(request))
		}
	}
	p.base.doEnd(newCtx, request, response, err, endTime)
}

func (p *PropagatingToDownstreamInstrumenter[REQUEST, RESPONSE]) StartAndEndWithOptions(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time, startOptions []trace.SpanStartOption, endOptions []trace.SpanEndOption) {
	newCtx := p.base.doStart(parentContext, request, startTime, startOptions...)
	if p.carrierGetter != nil {
		if p.prop != nil {
			p.prop.Inject(newCtx, p.carrierGetter(request))
		} else {
			otel.GetTextMapPropagator().Inject(newCtx, p.carrierGetter(request))
		}
	}
	p.base.doEnd(newCtx, request, response, err, endTime, endOptions...)
}

func (p *PropagatingToDownstreamInstrumenter[REQUEST, RESPONSE]) Start(parentContext context.Context, request REQUEST, options ...trace.SpanStartOption) context.Context {
	newCtx := p.base.Start(parentContext, request, options...)
	if p.carrierGetter != nil {
		if p.prop != nil {
			p.prop.Inject(newCtx, p.carrierGetter(request))
		} else {
			otel.GetTextMapPropagator().Inject(newCtx, p.carrierGetter(request))
		}

	}
	return newCtx
}

func (p *PropagatingToDownstreamInstrumenter[REQUEST, RESPONSE]) End(ctx context.Context, request REQUEST, response RESPONSE, err error, options ...trace.SpanEndOption) {
	p.base.End(ctx, request, response, err, options...)
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) ShouldStart(parentContext context.Context, request REQUEST) bool {
	return p.base.ShouldStart(parentContext, request)
}
func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) StartAndEnd(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time) {
	var ctx context.Context
	if p.carrierGetter != nil {
		var extracted context.Context
		if p.prop != nil {
			extracted = p.prop.Extract(parentContext, p.carrierGetter(request))
		} else {
			extracted = otel.GetTextMapPropagator().Extract(parentContext, p.carrierGetter(request))
		}
		ctx = p.base.doStart(extracted, request, startTime)
	} else {
		ctx = parentContext
	}
	p.base.doEnd(ctx, request, response, err, endTime)
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) StartAndEndWithOptions(parentContext context.Context, request REQUEST, response RESPONSE, err error, startTime, endTime time.Time, startOptions []trace.SpanStartOption, endOptions []trace.SpanEndOption) {
	var ctx context.Context
	if p.carrierGetter != nil {
		var extracted context.Context
		if p.prop != nil {
			extracted = p.prop.Extract(parentContext, p.carrierGetter(request))
		} else {
			extracted = otel.GetTextMapPropagator().Extract(parentContext, p.carrierGetter(request))
		}
		ctx = p.base.doStart(extracted, request, startTime, startOptions...)
	} else {
		ctx = parentContext
	}
	p.base.doEnd(ctx, request, response, err, endTime, endOptions...)
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) Start(parentContext context.Context, request REQUEST, options ...trace.SpanStartOption) context.Context {
	if p.carrierGetter != nil {
		var extracted context.Context
		if p.prop != nil {
			extracted = p.prop.Extract(parentContext, p.carrierGetter(request))
		} else {
			extracted = otel.GetTextMapPropagator().Extract(parentContext, p.carrierGetter(request))
		}
		return p.base.Start(extracted, request, options...)
	} else {
		return parentContext
	}
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) End(ctx context.Context, request REQUEST, response RESPONSE, err error, options ...trace.SpanEndOption) {
	p.base.End(ctx, request, response, err, options...)
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) StartWithTime(parentContext context.Context, request REQUEST, startTime time.Time, options ...trace.SpanStartOption) context.Context {
	if p.carrierGetter != nil {
		var extracted context.Context
		if p.prop != nil {
			extracted = p.prop.Extract(parentContext, p.carrierGetter(request))
		} else {
			extracted = otel.GetTextMapPropagator().Extract(parentContext, p.carrierGetter(request))
		}
		return p.base.doStart(extracted, request, startTime, options...)
	} else {
		return parentContext
	}
}

func (p *PropagatingFromUpstreamInstrumenter[REQUEST, RESPONSE]) EndWithTime(ctx context.Context, request REQUEST, response RESPONSE, err error, endTime time.Time, options ...trace.SpanEndOption) {
	p.base.doEnd(ctx, request, response, err, endTime, options...)
}
