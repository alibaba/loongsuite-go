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
	"errors"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/attribute"
)

type testRequest struct {
}

type testResponse struct {
	status string
}

type testNameExtractor struct {
}

func (t testNameExtractor) Extract(request testRequest) string {
	return "test"
}

type testOperationListener struct {
}

type disableEnabler struct {
}

func (d disableEnabler) Enable() bool {
	return false
}

type mockProp struct {
	val string
}

func (m *mockProp) Get(key string) string {
	return m.val
}

func (m *mockProp) Set(key string, value string) {
	m.val = value
}

func (m *mockProp) Keys() []string {
	return []string{"test"}
}

type myTextMapProp struct {
}

func (m *myTextMapProp) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	carrier.Set("test", "test")
}

func (m *myTextMapProp) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	t := carrier.Get("test")
	return context.WithValue(ctx, "test", t)
}

func (m *myTextMapProp) Fields() []string {
	return []string{"test"}
}

func (t *testOperationListener) OnBeforeStart(parentContext context.Context, startTimestamp time.Time) context.Context {
	return context.WithValue(parentContext, "startTs", startTimestamp)
}

func (t *testOperationListener) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTimestamp time.Time) context.Context {
	return context.WithValue(ctx, "startAttrs", startAttributes)
}

func (t *testOperationListener) OnAfterStart(context context.Context, endTimestamp time.Time) {
	if time.Now().Sub(endTimestamp).Seconds() > 5 {
		panic("duration too long")
	}
}

func (t *testOperationListener) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTimestamp time.Time) {
	if endAttributes[0].Key != "testAttribute" {
		panic("invalid attribute key")
	}
	if endAttributes[0].Value.AsString() != "testValue" {
		panic("invalid attribute value")
	}
}

type testAttributesExtractor struct {
}

func (t testAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request testRequest) ([]attribute.KeyValue, context.Context) {
	return []attribute.KeyValue{
		attribute.String("testAttribute", "testValue"),
	}, parentContext
}

func (t testAttributesExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request testRequest, response testResponse, err error) ([]attribute.KeyValue, context.Context) {
	return []attribute.KeyValue{
		attribute.String("testAttribute", "testValue"),
	}, context
}

type testContextCustomizer struct {
}

func (t testContextCustomizer) OnStart(ctx context.Context, request testRequest, startAttributes []attribute.KeyValue) context.Context {
	return context.WithValue(ctx, "test-customizer", "test-customizer")
}

func TestInstrumenter(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).AddContextCustomizers(testContextCustomizer{})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	newCtx := instrumenter.Start(ctx, testRequest{})
	if newCtx.Value("test-customizer") != "test-customizer" {
		t.Fatal("key test-customizer is not expected")
	}
	if newCtx.Value("startTs") == nil {
		t.Fatal("startTs is not expected")
	}
	if newCtx.Value("startAttrs") == nil {
		t.Fatal("startAttrs is not expected")
	}
	instrumenter.End(ctx, testRequest{}, testResponse{}, errors.New("abc"))
}

func TestStartAndEnd(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	instrumenter.StartAndEnd(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now())
	prop := mockProp{"test"}
	dsInstrumenter := builder.BuildPropagatingToDownstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	dsInstrumenter.StartAndEnd(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now())
	upInstrumenter := builder.BuildPropagatingFromUpstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	upInstrumenter.StartAndEnd(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now())
	// no panic here
}

func TestEnabler(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{}).
		SetInstrumentEnabler(disableEnabler{})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	newCtx := instrumenter.Start(ctx, testRequest{})
	if newCtx.Value("startTs") != nil {
		panic("the context should be an empty one")
	}
}

func TestPropFromUpStream(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{})
	prop := mockProp{"test"}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	instrumenter := builder.BuildPropagatingFromUpstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	ctx := context.Background()
	newCtx := instrumenter.Start(ctx, testRequest{})
	instrumenter.End(ctx, testRequest{}, testResponse{}, nil)
	if newCtx.Value("test") != "test" {
		panic("test attributes in context should be test")
	}
}

func TestPropToDownStream(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{})
	prop := mockProp{}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	instrumenter := builder.BuildPropagatingToDownstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	ctx := context.Background()
	instrumenter.Start(ctx, testRequest{})
	instrumenter.End(ctx, testRequest{}, testResponse{}, nil)
	if prop.val != "test" {
		panic("prop val should be test!")
	}
}

func TestStartAndEndWithOptions(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	instrumenter.StartAndEndWithOptions(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now(), nil, nil)
	prop := mockProp{"test"}
	dsInstrumenter := builder.BuildPropagatingToDownstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	dsInstrumenter.StartAndEndWithOptions(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now(), nil, nil)
	upInstrumenter := builder.BuildPropagatingFromUpstreamInstrumenter(func(request testRequest) propagation.TextMapCarrier {
		return &prop
	}, &myTextMapProp{})
	upInstrumenter.StartAndEndWithOptions(ctx, testRequest{}, testResponse{}, nil, time.Now(), time.Now(), nil, nil)
	// no panic here
}

func TestInstrumentationScope(t *testing.T) {
	builder := Builder[testRequest, testResponse]{}
	builder.Init().SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:      "test",
			Version:   "test",
			SchemaURL: "test",
		})
	ctx := context.Background()
	originalTP := otel.GetTracerProvider()
	traceProvider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(traceProvider)
	defer otel.SetTracerProvider(originalTP)
	instrumenter := builder.BuildInstrumenter()
	newCtx := instrumenter.Start(ctx, testRequest{})
	span := trace.SpanFromContext(newCtx)
	if readOnly, ok := span.(sdktrace.ReadOnlySpan); !ok {
		panic("it should be a readonly span")
	} else {
		if readOnly.InstrumentationScope().Name != "test" {
			panic("scope name should be test")
		}
		if readOnly.InstrumentationScope().Version != "test" {
			panic("scope version should be test")
		}
		if readOnly.InstrumentationScope().SchemaURL != "test" {
			panic("scope schema url should be test")
		}
	}
}

func TestSpanTimestamps(t *testing.T) {
	// The `startTime` and `endTime` of the generated span
	// must exactly match those in the input params of inst-api entry func.

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sr),
	)
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(originalTP)

	builder := Builder[testRequest, testResponse]{}
	builder.Init().
		SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		AddAttributesExtractor(testAttributesExtractor{}).
		AddOperationListeners(&testOperationListener{}).
		AddContextCustomizers(testContextCustomizer{})
	tracer := tp.
		Tracer("test-tracer")
	instrumenter := builder.BuildInstrumenterWithTracer(tracer)
	ctx := context.Background()
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Second)
	instrumenter.StartAndEnd(ctx, testRequest{}, testResponse{}, nil, startTime, endTime)
	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("no spans captured")
	}
	recordedSpan := spans[0]
	assert.Equal(t, startTime, recordedSpan.StartTime())
	assert.Equal(t, endTime, recordedSpan.EndTime())
}

// SpanRecorder records started and ended spans.
type SpanRecorder struct {
	started []sdktrace.ReadWriteSpan
	ended   []sdktrace.ReadOnlySpan
}
