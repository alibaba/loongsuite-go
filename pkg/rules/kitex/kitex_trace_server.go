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

package kitex

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	sdktrace "go.opentelemetry.io/otel/trace"
)

var kitexServerInstrumenter = BuildKitexServerInstrumenter()

type serverTracer struct{}

func (s *serverTracer) Start(ctx context.Context) context.Context {
	tc := &TraceCarrier{}
	return WithTraceCarrier(ctx, tc)
}

func (s *serverTracer) Finish(ctx context.Context) {
	tc := TraceCarrierFromContext(ctx)
	if tc == nil {
		return
	}
	// set stack and error here, thus kitex's panic stack is a interface
	span := tc.Span()
	ri := rpcinfo.GetRPCInfo(ctx)
	panicMsg, panicStack, err := parseRPCError(ri)
	if err != nil {
		opts := make([]sdktrace.EventOption, 0)
		if span == nil || !span.IsRecording() {
			return
		}
		opts = append(opts, sdktrace.WithAttributes(
			semconv.ExceptionType(panicMsg),
			semconv.ExceptionMessage(err.Error()),
			semconv.ExceptionStacktrace(panicStack),
		))
		span.AddEvent(semconv.ExceptionEventName, opts...)
		ctx = sdktrace.ContextWithSpan(ctx, span)
	}
	kitexServerInstrumenter.End(ctx, ri, ri, nil)
}

func ServerMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			tc := TraceCarrierFromContext(ctx)
			if tc == nil {
				return next(ctx, req, resp)
			}
			md := metainfo.GetAllValues(ctx)
			if md == nil {
				md = make(map[string]string)
			}
			grpcMeta, ok := metadata.FromIncomingContext(ctx)
			if ok {
				for k1, v1 := range grpcMeta {
					if len(v1) > 0 {
						md[k1] = v1[0]
					}
				}
			}
			ctx = Extract(ctx, md)
			ri := rpcinfo.GetRPCInfo(ctx)
			ctx = kitexServerInstrumenter.Start(ctx, ri)
			tc.SetSpan(sdktrace.SpanFromContext(ctx))
			return next(ctx, req, resp)
		}
	}
}
