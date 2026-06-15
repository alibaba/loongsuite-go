// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package rules

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oTrace "go.opentelemetry.io/otel/trace"

	_ "unsafe"
)

//go:linkname consumerMessageOnEnter kafka-demo/consumer.consumerMessageOnEnter
func consumerMessageOnEnter(call api.CallContext, k interface{}, ctx context.Context, message kafka.Message) {
	for _, v := range message.Headers {
		if v.Key == "traceparent" {
			var headerMap propagation.MapCarrier
			headerMap = make(map[string]string)
			headerMap[v.Key] = string(v.Value)
			ctx = otel.GetTextMapPropagator().Extract(ctx, headerMap)
			tracer := otel.GetTracerProvider().Tracer("")
			opts := append([]oTrace.SpanStartOption{}, oTrace.WithSpanKind(oTrace.SpanKindConsumer))
			_, span := tracer.Start(ctx, "consumer message", opts...)
			temp := make(map[string]interface{}, 1)
			temp["span"] = span
			call.SetData(temp)
			break
		}
	}
}

//go:linkname consumerMessageOnExit kafka-demo/consumer.consumerMessageOnExit
func consumerMessageOnExit(call api.CallContext, err error) {
	if call.GetData() == nil {
		return
	}
	temp := call.GetData().(map[string]interface{})
	span := temp["span"].(oTrace.Span)
	if span != nil {
		span.End()
		return
	}
}
