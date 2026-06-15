// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type Filter func(*InterceptorInfo) bool

// grpcOtelConfig is a group of options for this instrumentation.
type grpcOtelConfig struct {
	Filter           Filter
	Propagators      propagation.TextMapPropagator
	TracerProvider   trace.TracerProvider
	SpanStartOptions []trace.SpanStartOption

	ReceivedEvent bool
	SentEvent     bool

	tracer trace.Tracer

	DestId string
}

// Option applies an option value for a grpcOtelConfig.
type Option interface {
	apply(*grpcOtelConfig)
}

func (c *grpcOtelConfig) handleRPC(ctx context.Context, rs stats.RPCStats, isServer bool) { // nolint: revive  // isServer is not a control flag.
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}
	var (
		messageId int64
	)
	gCtx, _ := ctx.Value(gRPCContextKey{}).(*gRPCContext)
	switch rs := rs.(type) {
	case *stats.Begin:
	case *stats.InPayload:
		if c.ReceivedEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeReceived,
					semconv.MessageIDKey.Int64(messageId),
					semconv.MessageUncompressedSizeKey.Int(rs.Length),
				),
			)
		}
	case *stats.OutPayload:
		if c.SentEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeSent,
					semconv.MessageIDKey.Int64(messageId),
					//semconv.MessageCompressedSizeKey.Int(rs.CompressedLength),
					semconv.MessageUncompressedSizeKey.Int(rs.Length),
				),
			)
		}
	case *stats.OutTrailer:
	case *stats.End:
		if rs.Error != nil {
			s, _ := status.FromError(rs.Error)
			if isServer {
				grpcServerInstrument.End(ctx, grpcRequest{}, grpcResponse{
					statusCode: int(s.Code()),
				}, rs.Error)
			} else {
				grpcClientInstrument.End(ctx, grpcRequest{}, grpcResponse{
					statusCode: int(s.Code()),
				}, rs.Error)
			}

		} else {
			methodName := ""
			if gCtx != nil {
				methodName = gCtx.methodName
			}
			if isServer {
				grpcServerInstrument.End(ctx, grpcRequest{
					methodName: methodName,
				}, grpcResponse{
					statusCode: 200,
				}, nil)
			} else {
				grpcClientInstrument.End(ctx, grpcRequest{
					methodName: methodName,
				}, grpcResponse{
					statusCode: 200,
				}, nil)
			}

		}
	default:
		return
	}
}

// newConfig returns a grpcOtelConfig configured with all the passed Options.
func newConfig(opts []Option, role string) *grpcOtelConfig {
	c := &grpcOtelConfig{
		Propagators: otel.GetTextMapPropagator(),
	}
	for _, o := range opts {
		o.apply(c)
	}

	return c
}

type propagatorsOption struct{ p propagation.TextMapPropagator }

func (o propagatorsOption) apply(c *grpcOtelConfig) {
	if o.p != nil {
		c.Propagators = o.p
	}
}

// WithPropagators returns an Option to use the Propagators when extracting
// and injecting trace context from requests.
func WithPropagators(p propagation.TextMapPropagator) Option {
	return propagatorsOption{p: p}
}

type tracerProviderOption struct{ tp trace.TracerProvider }

func (o tracerProviderOption) apply(c *grpcOtelConfig) {
	if o.tp != nil {
		c.TracerProvider = o.tp
	}
}

// WithInterceptorFilter returns an Option to use the request filter.
//
// Deprecated: Use stats handlers instead.
func WithInterceptorFilter(f Filter) Option {
	return interceptorFilterOption{f: f}
}

type interceptorFilterOption struct {
	f Filter
}

func (o interceptorFilterOption) apply(c *grpcOtelConfig) {
	if o.f != nil {
		c.Filter = o.f
	}
}

// WithTracerProvider returns an Option to use the TracerProvider when
// creating a Tracer.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return tracerProviderOption{tp: tp}
}

//type meterProviderOption struct{ mp metric.MeterProvider }

/*func (o meterProviderOption) apply(c *grpcOtelConfig) {
	if o.mp != nil {
		c.MeterProvider = o.mp
	}
}

// WithMeterProvider returns an Option to use the MeterProvider when
// creating a Meter. If this option is not provide the global MeterProvider will be used.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return meterProviderOption{mp: mp}
}*/

// Event type that can be recorded, see WithMessageEvents.
type Event int

// Different types of events that can be recorded, see WithMessageEvents.
const (
	ReceivedEvents Event = iota
	SentEvents
)

type messageEventsProviderOption struct {
	events []Event
}

func (m messageEventsProviderOption) apply(c *grpcOtelConfig) {
	for _, e := range m.events {
		switch e {
		case ReceivedEvents:
			c.ReceivedEvent = true
		case SentEvents:
			c.SentEvent = true
		}
	}
}

// WithMessageEvents configures the Handler to record the specified events
// (span.AddEvent) on spans. By default only summary attributes are added at the
// end of the request.
//
// Valid events are:
//   - ReceivedEvents: Record the number of bytes read after every gRPC read operation.
//   - SentEvents: Record the number of bytes written after every gRPC write operation.
func WithMessageEvents(events ...Event) Option {
	return messageEventsProviderOption{events: events}
}

type spanStartOption struct{ opts []trace.SpanStartOption }

func (o spanStartOption) apply(c *grpcOtelConfig) {
	c.SpanStartOptions = append(c.SpanStartOptions, o.opts...)
}

// WithSpanOptions configures an additional set of
// trace.SpanOptions, which are applied to each new span.
func WithSpanOptions(opts ...trace.SpanStartOption) Option {
	return spanStartOption{opts}
}
