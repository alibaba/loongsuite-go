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

package kafka

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"os"
)

// Instrumentation enabler controller
var kafkaEnabler = kafkaInnerEnabler{os.Getenv("OTEL_SEGMENTIO_KAFKA_ENABLED") != "false"}

// Cache Instrumenter instances to avoid repeated creation
var (
	producerInstrumenter = buildKafkaProducerInstrumenter()
	consumerInstrumenter = buildKafkaConsumerInstrumenter()
)

type kafkaInnerEnabler struct {
	enabled bool
}

func (k kafkaInnerEnabler) Enable() bool {
	return k.enabled
}

// KafkaProducerCarrier implements OpenTelemetry propagator carrier interface for producers
type kafkaProducerCarrier struct {
	messages []*kafka.Message
}

func (carrier kafkaProducerCarrier) Get(key string) string {
	return ""
}

func (carrier kafkaProducerCarrier) Set(key, value string) {
	for _, message := range carrier.messages {
		message.Headers = append(message.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}
}

func (carrier kafkaProducerCarrier) Keys() []string {
	return []string{}
}

// KafkaConsumerCarrier implements OpenTelemetry propagator carrier interface for consumers
type kafkaConsumerCarrier struct {
	message kafka.Message
}

func (carrier kafkaConsumerCarrier) Get(key string) string {
	if carrier.message.Headers != nil {
		for _, header := range carrier.message.Headers {
			if header.Key == key {
				return string(header.Value)
			}
		}
	}
	return ""
}

func (carrier kafkaConsumerCarrier) Set(key, value string) {
	// Consumer carrier doesn't need to implement Set method
}

func (carrier kafkaConsumerCarrier) Keys() []string {
	return []string{}
}

// KafkaProducerStatusExtractor extracts producer operation status
type kafkaProducerStatusExtractor struct {
}

func (extractor *kafkaProducerStatusExtractor) Extract(span trace.Span, request kafkaProducerReq, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// KafkaConsumerStatusExtractor extracts consumer operation status
type kafkaConsumerStatusExtractor struct{}

func (extractor *kafkaConsumerStatusExtractor) Extract(span trace.Span, request kafkaConsumerReq, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// KafkaMessageProducerAttributesGetter retrieves producer message attributes
type kafkaMessageProducerAttrsGetter struct{}

func (getter kafkaMessageProducerAttrsGetter) IsAnonymousDestination(request kafkaProducerReq) bool {
	return false
}

func (getter kafkaMessageProducerAttrsGetter) GetDestinationPartitionId(request kafkaProducerReq) string {
	return ""
}

func (getter kafkaMessageProducerAttrsGetter) GetSystem(request kafkaProducerReq) string {
	return "kafka"
}

func (getter kafkaMessageProducerAttrsGetter) GetDestination(request kafkaProducerReq) string {
	return request.topic
}

func (getter kafkaMessageProducerAttrsGetter) GetDestinationTemplate(request kafkaProducerReq) string {
	return ""
}

func (getter kafkaMessageProducerAttrsGetter) IsTemporaryDestination(request kafkaProducerReq) bool {
	return false
}

func (getter kafkaMessageProducerAttrsGetter) isAnonymousDestination(request kafkaProducerReq) bool {
	return false
}

func (getter kafkaMessageProducerAttrsGetter) GetConversationId(request kafkaProducerReq) string {
	return ""
}

func (getter kafkaMessageProducerAttrsGetter) GetMessageBodySize(request kafkaProducerReq) int64 {
	return 0
}

func (getter kafkaMessageProducerAttrsGetter) GetMessageEnvelopSize(request kafkaProducerReq) int64 {
	return 0
}

func (getter kafkaMessageProducerAttrsGetter) GetMessageId(request kafkaProducerReq, response any) string {
	return ""
}

func (getter kafkaMessageProducerAttrsGetter) GetClientId(request kafkaProducerReq) string {
	return ""
}

func (getter kafkaMessageProducerAttrsGetter) GetBatchMessageCount(request kafkaProducerReq, response any) int64 {
	return int64(len(request.msgs))
}

func (getter kafkaMessageProducerAttrsGetter) GetMessageHeader(request kafkaProducerReq, name string) []string {
	return []string{}
}

// KafkaMessageConsumerAttributesGetter retrieves consumer message attributes
type kafkaMessageConsumerAttrsGetter struct{}

func (getter kafkaMessageConsumerAttrsGetter) IsAnonymousDestination(request kafkaConsumerReq) bool {
	return false
}

func (getter kafkaMessageConsumerAttrsGetter) GetDestinationPartitionId(request kafkaConsumerReq) string {
	return ""
}

func (getter kafkaMessageConsumerAttrsGetter) GetSystem(request kafkaConsumerReq) string {
	return "kafka"
}

func (getter kafkaMessageConsumerAttrsGetter) GetDestination(request kafkaConsumerReq) string {
	return request.msg.Topic
}

func (getter kafkaMessageConsumerAttrsGetter) GetDestinationTemplate(request kafkaConsumerReq) string {
	return ""
}

func (getter kafkaMessageConsumerAttrsGetter) IsTemporaryDestination(request kafkaConsumerReq) bool {
	return false
}

func (getter kafkaMessageConsumerAttrsGetter) isAnonymousDestination(request kafkaConsumerReq) bool {
	return false
}

func (getter kafkaMessageConsumerAttrsGetter) GetConversationId(request kafkaConsumerReq) string {
	return ""
}

func (getter kafkaMessageConsumerAttrsGetter) GetMessageBodySize(request kafkaConsumerReq) int64 {
	// Calculate message body size: 4(topic length) + 1(compression type) + 1(attributes) + 4(key length) + key length + 4(value length) + value length + 8(timestamp)
	return int64(4 + 1 + 1 + 4 + len(request.msg.Key) + 4 + len(request.msg.Value) + 8)
}

func (getter kafkaMessageConsumerAttrsGetter) GetMessageEnvelopSize(request kafkaConsumerReq) int64 {
	return 0
}

func (getter kafkaMessageConsumerAttrsGetter) GetMessageId(request kafkaConsumerReq, response any) string {
	return ""
}

func (getter kafkaMessageConsumerAttrsGetter) GetClientId(request kafkaConsumerReq) string {
	return ""
}

func (getter kafkaMessageConsumerAttrsGetter) GetBatchMessageCount(request kafkaConsumerReq, response any) int64 {
	return 1
}

func (getter kafkaMessageConsumerAttrsGetter) GetMessageHeader(request kafkaConsumerReq, name string) []string {
	if request.msg.Headers == nil {
		return nil
	}
	var headerValues []string
	for _, header := range request.msg.Headers {
		if header.Key == name {
			headerValues = append(headerValues, string(header.Value))
		}
	}
	return headerValues
}

// KafkaProducerAttributesExtractor extracts producer attributes
type kafkaProducerAttributesExtractor struct {
}

func (extractor *kafkaProducerAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request kafkaProducerReq) ([]attribute.KeyValue, context.Context) {
	kafkaAttributes := []attribute.KeyValue{
		semconv.MessagingSystemKafka,
		semconv.MessagingDestinationNameKey.String(request.topic),
		semconv.MessagingOperationName("publish"),
	}
	return append(attributes, kafkaAttributes...), parentContext
}

func (extractor *kafkaProducerAttributesExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request kafkaProducerReq, response any, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, ctx
}

// KafkaConsumerAttributesExtractor extracts consumer attributes
type kafkaConsumerAttributesExtractor struct {
}

func (extractor *kafkaConsumerAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request kafkaConsumerReq) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (extractor *kafkaConsumerAttributesExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request kafkaConsumerReq, response any, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, ctx
}

// Build Kafka producer instrumenter
func buildKafkaProducerInstrumenter() instrumenter.Instrumenter[kafkaProducerReq, any] {
	builder := instrumenter.Builder[kafkaProducerReq, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.KAFKAGO_PRODUCER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[kafkaProducerReq, any]{
			Getter:        kafkaMessageProducerAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[kafkaProducerReq]{}).
		SetSpanStatusExtractor(&kafkaProducerStatusExtractor{}).
		AddAttributesExtractor(&kafkaProducerAttributesExtractor{}).
		BuildPropagatingToDownstreamInstrumenter(
			func(request kafkaProducerReq) propagation.TextMapCarrier {
				return kafkaProducerCarrier{messages: request.msgs}
			},
			otel.GetTextMapPropagator(),
		)
}

// Build Kafka consumer instrumenter
func buildKafkaConsumerInstrumenter() instrumenter.Instrumenter[kafkaConsumerReq, any] {
	builder := instrumenter.Builder[kafkaConsumerReq, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.KAFKAGO_CONSUMER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[kafkaConsumerReq, any]{
			Getter:        kafkaMessageConsumerAttrsGetter{},
			OperationName: message.PROCESS,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[kafkaConsumerReq]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[kafkaConsumerReq, any, kafkaMessageConsumerAttrsGetter]{
			Operation: message.PROCESS,
		}).
		AddAttributesExtractor(&kafkaConsumerAttributesExtractor{}).
		BuildPropagatingFromUpstreamInstrumenter(
			func(request kafkaConsumerReq) propagation.TextMapCarrier {
				return kafkaConsumerCarrier{message: request.msg}
			},
			otel.GetTextMapPropagator(),
		)
}
