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

package amqp091

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

//go:linkname consumeOnEnter github.com/rabbitmq/amqp091-go.consumeOnEnter
func consumeOnEnter(call api.CallContext,
	_ interface{},
	tag string,
	msg *amqp.Delivery,
) {
	request := RabbitRequest{
		operationName:   "receive",
		destinationName: msg.Exchange + ":" + msg.RoutingKey,
		messageId:       msg.MessageId,
		bodySize:        int64(len(msg.Body)),
		conversationID:  msg.CorrelationId,
		headers:         msg.Headers,
	}
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		semconv.MessagingRabbitmqDestinationRoutingKey(msg.RoutingKey),
		semconv.MessagingRabbitmqMessageDeliveryTag(int(msg.DeliveryTag)), attribute.KeyValue{
			Key:   semconv.MessagingOperationTypeKey,
			Value: attribute.StringValue(request.operationName),
		}, attribute.KeyValue{
			Key:   "messaging.rabbitmq.message.consumer_tag",
			Value: attribute.StringValue(msg.ConsumerTag),
		},
	)
	ctx = RabbitMQConsumeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["rabbitMQ_consume_request"] = request
	call.SetData(data)
}

//go:linkname consumeOnExit github.com/rabbitmq/amqp091-go.consumeOnExit
func consumeOnExit(call api.CallContext, b bool) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	request, ok := data["rabbitMQ_consume_request"].(RabbitRequest)
	if !ok {
		return
	}
	RabbitMQConsumeInstrumenter.End(ctx, request, nil, nil)
}
