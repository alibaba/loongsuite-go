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

package sarama

import (
	"context"
	"encoding/binary"
	"os"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

var saramaEnabler = saramaInnerEnabler{os.Getenv("OTEL_SHOPIFY_SARAMA_ENABLED") != "false"}

var (
	producerInstrumenter = buildProducerInstrumenter()
	consumerInstrumenter = buildConsumerInstrumenter()
)

type saramaInnerEnabler struct {
	enabled bool
}

func (e saramaInnerEnabler) Enable() bool {
	return e.enabled
}

// --- Producer carrier ---

type saramaProducerCarrier struct {
	msg        *sarama.ProducerMessage
	msgVersion sarama.KafkaVersion
}

var _ propagation.TextMapCarrier = saramaProducerCarrier{}

func (c saramaProducerCarrier) Get(key string) string {
	if c.msg == nil {
		return ""
	}
	for _, h := range c.msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c saramaProducerCarrier) Set(key, val string) {
	if c.msg == nil {
		return
	}
	if !c.msgVersion.IsAtLeast(sarama.V0_11_0_0) {
		return
	}
	for i := 0; i < len(c.msg.Headers); i++ {
		if string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(val),
	})
}

func (c saramaProducerCarrier) Keys() []string {
	if c.msg == nil {
		return nil
	}
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}

// --- Consumer carrier ---

type saramaConsumerCarrier struct {
	msg *sarama.ConsumerMessage
}

var _ propagation.TextMapCarrier = saramaConsumerCarrier{}

func (c saramaConsumerCarrier) Get(key string) string {
	if c.msg == nil {
		return ""
	}
	for _, h := range c.msg.Headers {
		if h != nil && string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c saramaConsumerCarrier) Set(key, val string) {
	if c.msg == nil {
		return
	}
	for i := 0; i < len(c.msg.Headers); i++ {
		if c.msg.Headers[i] != nil && string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, &sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(val),
	})
}

func (c saramaConsumerCarrier) Keys() []string {
	if c.msg == nil {
		return nil
	}
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}

// --- Status extractors ---

type producerStatusExtractor struct{}

func (e *producerStatusExtractor) Extract(span trace.Span, request saramaProducerRequest, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type consumerStatusExtractor struct{}

func (e *consumerStatusExtractor) Extract(span trace.Span, request saramaConsumerRequest, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// --- Producer attrs getter (implements message.MessageAttrsGetter) ---

type producerAttrsGetter struct{}

func (g producerAttrsGetter) GetSystem(request saramaProducerRequest) string         { return "kafka" }
func (g producerAttrsGetter) GetDestination(request saramaProducerRequest) string {
	if request.msg != nil {
		return request.msg.Topic
	}
	return "unknown"
}
func (g producerAttrsGetter) GetDestinationTemplate(request saramaProducerRequest) string { return "" }
func (g producerAttrsGetter) IsTemporaryDestination(request saramaProducerRequest) bool   { return false }
func (g producerAttrsGetter) IsAnonymousDestination(request saramaProducerRequest) bool   { return false }
func (g producerAttrsGetter) GetConversationId(request saramaProducerRequest) string      { return "" }
func (g producerAttrsGetter) GetMessageBodySize(request saramaProducerRequest) int64 {
	return int64(producerMsgPayloadSize(request.msg, request.msgVersion))
}
func (g producerAttrsGetter) GetMessageEnvelopSize(request saramaProducerRequest) int64 { return 0 }
func (g producerAttrsGetter) GetMessageId(request saramaProducerRequest, response any) string {
	if resp, ok := response.(saramaProducerResponse); ok && resp.offset != nil {
		return *resp.offset
	}
	return ""
}
func (g producerAttrsGetter) GetClientId(request saramaProducerRequest) string { return "" }
func (g producerAttrsGetter) GetBatchMessageCount(request saramaProducerRequest, response any) int64 {
	return 1
}
func (g producerAttrsGetter) GetMessageHeader(request saramaProducerRequest, name string) []string {
	return nil
}
func (g producerAttrsGetter) GetDestinationPartitionId(request saramaProducerRequest) string {
	return ""
}

// --- Consumer attrs getter (implements message.MessageAttrsGetter) ---

type consumerAttrsGetter struct{}

func (g consumerAttrsGetter) GetSystem(request saramaConsumerRequest) string { return "kafka" }
func (g consumerAttrsGetter) GetDestination(request saramaConsumerRequest) string {
	if request.msg != nil {
		return request.msg.Topic
	}
	return "unknown"
}
func (g consumerAttrsGetter) GetDestinationTemplate(request saramaConsumerRequest) string { return "" }
func (g consumerAttrsGetter) IsTemporaryDestination(request saramaConsumerRequest) bool   { return false }
func (g consumerAttrsGetter) IsAnonymousDestination(request saramaConsumerRequest) bool   { return false }
func (g consumerAttrsGetter) GetConversationId(request saramaConsumerRequest) string      { return "" }
func (g consumerAttrsGetter) GetMessageBodySize(request saramaConsumerRequest) int64 {
	return int64(consumerMsgPayloadSize(request.msg))
}
func (g consumerAttrsGetter) GetMessageEnvelopSize(request saramaConsumerRequest) int64 { return 0 }
func (g consumerAttrsGetter) GetMessageId(request saramaConsumerRequest, response any) string {
	if request.offset != nil {
		return *request.offset
	}
	return ""
}
func (g consumerAttrsGetter) GetClientId(request saramaConsumerRequest) string { return "" }
func (g consumerAttrsGetter) GetBatchMessageCount(request saramaConsumerRequest, response any) int64 {
	return 1
}
func (g consumerAttrsGetter) GetMessageHeader(request saramaConsumerRequest, name string) []string {
	if request.msg == nil {
		return nil
	}
	var vals []string
	for _, h := range request.msg.Headers {
		if h != nil && string(h.Key) == name {
			vals = append(vals, string(h.Value))
		}
	}
	return vals
}
func (g consumerAttrsGetter) GetDestinationPartitionId(request saramaConsumerRequest) string {
	if request.partition != nil {
		return strconv.Itoa(*request.partition)
	}
	return ""
}

// --- Additional attributes extractors ---

type producerExtraAttrsExtractor struct{}

func (e *producerExtraAttrsExtractor) OnStart(attrs []attribute.KeyValue, parentContext context.Context, request saramaProducerRequest) ([]attribute.KeyValue, context.Context) {
	return attrs, parentContext
}

func (e *producerExtraAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, request saramaProducerRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	if resp, ok := response.(saramaProducerResponse); ok && resp.partition != nil {
		attrs = append(attrs, semconv.MessagingDestinationPartitionIDKey.String(strconv.Itoa(*resp.partition)))
	}
	return attrs, ctx
}

type consumerExtraAttrsExtractor struct{}

func (e *consumerExtraAttrsExtractor) OnStart(attrs []attribute.KeyValue, parentContext context.Context, request saramaConsumerRequest) ([]attribute.KeyValue, context.Context) {
	return attrs, parentContext
}

func (e *consumerExtraAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, request saramaConsumerRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	if request.partition != nil {
		attrs = append(attrs, semconv.MessagingDestinationPartitionIDKey.String(strconv.Itoa(*request.partition)))
	}
	return attrs, ctx
}

// --- Message payload size helpers ---

func producerMsgPayloadSize(msg *sarama.ProducerMessage, kafkaVersion sarama.KafkaVersion) int {
	if msg == nil {
		return 0
	}
	maximumRecordOverhead := 5*binary.MaxVarintLen32 + binary.MaxVarintLen64 + 1
	producerMessageOverhead := 26
	v := 1
	if kafkaVersion.IsAtLeast(sarama.V0_11_0_0) {
		v = 2
	}
	size := producerMessageOverhead
	if v >= 2 {
		size = maximumRecordOverhead
		for _, h := range msg.Headers {
			size += len(h.Key) + len(h.Value) + 2*binary.MaxVarintLen32
		}
	}
	if msg.Key != nil {
		size += msg.Key.Length()
	}
	if msg.Value != nil {
		size += msg.Value.Length()
	}
	return size
}

func consumerMsgPayloadSize(msg *sarama.ConsumerMessage) int {
	if msg == nil {
		return 0
	}
	size := 5*binary.MaxVarintLen32 + binary.MaxVarintLen64 + 1
	for _, h := range msg.Headers {
		if h != nil {
			size += len(h.Key) + len(h.Value) + 2*binary.MaxVarintLen32
		}
	}
	if msg.Key != nil {
		size += len(msg.Key)
	}
	if msg.Value != nil {
		size += len(msg.Value)
	}
	return size
}

// --- Instrumenter builders ---

func buildProducerInstrumenter() instrumenter.Instrumenter[saramaProducerRequest, any] {
	builder := instrumenter.Builder[saramaProducerRequest, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SARAMA_PRODUCER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[saramaProducerRequest, any]{
			Getter:        producerAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[saramaProducerRequest]{}).
		SetSpanStatusExtractor(&producerStatusExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[saramaProducerRequest, any, producerAttrsGetter]{
			Operation: message.PUBLISH,
		}).
		AddAttributesExtractor(&producerExtraAttrsExtractor{}).
		BuildPropagatingToDownstreamInstrumenter(
			func(request saramaProducerRequest) propagation.TextMapCarrier {
				if request.msg == nil {
					return nil
				}
				return saramaProducerCarrier{msg: request.msg, msgVersion: request.msgVersion}
			},
			otel.GetTextMapPropagator(),
		)
}

func buildConsumerInstrumenter() instrumenter.Instrumenter[saramaConsumerRequest, any] {
	builder := instrumenter.Builder[saramaConsumerRequest, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SARAMA_CONSUMER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[saramaConsumerRequest, any]{
			Getter:        consumerAttrsGetter{},
			OperationName: message.RECEIVE,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[saramaConsumerRequest]{}).
		SetSpanStatusExtractor(&consumerStatusExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[saramaConsumerRequest, any, consumerAttrsGetter]{
			Operation: message.RECEIVE,
		}).
		AddAttributesExtractor(&consumerExtraAttrsExtractor{}).
		BuildPropagatingFromUpstreamInstrumenter(
			func(request saramaConsumerRequest) propagation.TextMapCarrier {
				if request.msg == nil {
					return nil
				}
				return saramaConsumerCarrier{msg: request.msg}
			},
			otel.GetTextMapPropagator(),
		)
}
