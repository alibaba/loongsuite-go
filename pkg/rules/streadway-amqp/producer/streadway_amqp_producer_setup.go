// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package producer

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname publishChannelOnEnter github.com/streadway/amqp.publishChannelOnEnter
func publishChannelOnEnter(call api.CallContext, ch *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) {
	if ch == nil {
		return
	}

	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindProducer)}
	ctx, span := otel.Tracer("github.com/streadway/amqp").Start(context.Background(), exchange+" publish", opts...)
	span.SetAttributes(
		semconv.MessagingSystemRabbitmq,
		semconv.MessagingDestinationName(exchange),
		attribute.String("messaging.operation.name", "publish"),
		semconv.MessagingRabbitmqDestinationRoutingKey(key),
	)

	// Keep behavior compatible with historical cascading/non-cascading tests:
	// only propagate when caller provided message headers.
	if msg.Headers != nil {
		otel.GetTextMapPropagator().Inject(ctx, amqpProducerTextMapCarrier{headers: msg.Headers})
		call.SetParam(5, msg)
	}

	call.SetData(map[string]interface{}{
		"span": span,
	})
}

//go:linkname publishChannelOnExit github.com/streadway/amqp.publishChannelOnExit
func publishChannelOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	span, _ := data["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	defer span.End()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}
