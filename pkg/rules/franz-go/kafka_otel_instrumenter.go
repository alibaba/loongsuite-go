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

package franz_go

import (
	"context"
	"os"
	"strconv"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

var franzEnabler = franzInnerEnabler{os.Getenv("OTEL_FRANZ_GO_ENABLED") != "false"}

var (
	franzKafkaProducerInstrumenter = buildFranzKafkaProducerInstrumenter()
	franzKafkaConsumerInstrumenter = buildFranzKafkaConsumerInstrumenter()
)

type franzInnerEnabler struct {
	enabled bool
}

func (e franzInnerEnabler) Enable() bool {
	return e.enabled
}

// --- Producer ---

type franzProducerStatusExtractor struct{}

func (e *franzProducerStatusExtractor) Extract(span trace.Span, request franzKafkaProducerRequest, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type franzProducerAttrsGetter struct{}

func (g franzProducerAttrsGetter) GetSystem(request franzKafkaProducerRequest) string {
	return "kafka"
}

func (g franzProducerAttrsGetter) GetDestination(request franzKafkaProducerRequest) string {
	return request.topic
}

func (g franzProducerAttrsGetter) GetDestinationTemplate(request franzKafkaProducerRequest) string {
	return ""
}

func (g franzProducerAttrsGetter) IsTemporaryDestination(request franzKafkaProducerRequest) bool {
	return false
}

func (g franzProducerAttrsGetter) IsAnonymousDestination(request franzKafkaProducerRequest) bool {
	return false
}

func (g franzProducerAttrsGetter) GetConversationId(request franzKafkaProducerRequest) string {
	return ""
}

func (g franzProducerAttrsGetter) GetMessageBodySize(request franzKafkaProducerRequest) int64 {
	if request.Record != nil {
		return int64(len(request.Record.Value))
	}
	return 0
}

func (g franzProducerAttrsGetter) GetMessageEnvelopSize(request franzKafkaProducerRequest) int64 {
	return 0
}

func (g franzProducerAttrsGetter) GetMessageId(request franzKafkaProducerRequest, response any) string {
	if resp, ok := response.(franzKafkaProducerResponse); ok && resp.offset != nil {
		return *resp.offset
	}
	return ""
}

func (g franzProducerAttrsGetter) GetClientId(request franzKafkaProducerRequest) string {
	return ""
}

func (g franzProducerAttrsGetter) GetBatchMessageCount(request franzKafkaProducerRequest, response any) int64 {
	return 1
}

func (g franzProducerAttrsGetter) GetMessageHeader(request franzKafkaProducerRequest, name string) []string {
	if request.Record == nil {
		return nil
	}
	for _, h := range request.Record.Headers {
		if h.Key == name {
			return []string{string(h.Value)}
		}
	}
	return nil
}

func (g franzProducerAttrsGetter) GetDestinationPartitionId(request franzKafkaProducerRequest) string {
	if request.Record != nil {
		return strconv.Itoa(int(request.Record.Partition))
	}
	return ""
}

type franzProducerAttrsExtractor struct{}

func (e *franzProducerAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request franzKafkaProducerRequest) ([]attribute.KeyValue, context.Context) {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKafka,
		semconv.MessagingDestinationNameKey.String(request.topic),
		semconv.MessagingOperationName("publish"),
	}
	return append(attributes, attrs...), parentContext
}

func (e *franzProducerAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request franzKafkaProducerRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, ctx
}

// --- Consumer ---

type franzConsumerStatusExtractor struct{}

func (e *franzConsumerStatusExtractor) Extract(span trace.Span, request franzKafkaConsumerRequest, response any, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type franzConsumerAttrsGetter struct{}

func (g franzConsumerAttrsGetter) GetSystem(request franzKafkaConsumerRequest) string {
	return "kafka"
}

func (g franzConsumerAttrsGetter) GetDestination(request franzKafkaConsumerRequest) string {
	return request.topic
}

func (g franzConsumerAttrsGetter) GetDestinationTemplate(request franzKafkaConsumerRequest) string {
	return ""
}

func (g franzConsumerAttrsGetter) IsTemporaryDestination(request franzKafkaConsumerRequest) bool {
	return false
}

func (g franzConsumerAttrsGetter) IsAnonymousDestination(request franzKafkaConsumerRequest) bool {
	return false
}

func (g franzConsumerAttrsGetter) GetConversationId(request franzKafkaConsumerRequest) string {
	return ""
}

func (g franzConsumerAttrsGetter) GetMessageBodySize(request franzKafkaConsumerRequest) int64 {
	if request.Record != nil {
		return int64(len(request.Record.Value))
	}
	return 0
}

func (g franzConsumerAttrsGetter) GetMessageEnvelopSize(request franzKafkaConsumerRequest) int64 {
	return 0
}

func (g franzConsumerAttrsGetter) GetMessageId(request franzKafkaConsumerRequest, response any) string {
	if request.Record != nil {
		return strconv.FormatInt(request.Record.Offset, 10)
	}
	return ""
}

func (g franzConsumerAttrsGetter) GetClientId(request franzKafkaConsumerRequest) string {
	return ""
}

func (g franzConsumerAttrsGetter) GetBatchMessageCount(request franzKafkaConsumerRequest, response any) int64 {
	return 1
}

func (g franzConsumerAttrsGetter) GetMessageHeader(request franzKafkaConsumerRequest, name string) []string {
	if request.Record == nil {
		return nil
	}
	for _, h := range request.Record.Headers {
		if h.Key == name {
			return []string{string(h.Value)}
		}
	}
	return nil
}

func (g franzConsumerAttrsGetter) GetDestinationPartitionId(request franzKafkaConsumerRequest) string {
	if request.Record != nil {
		return strconv.Itoa(int(request.Record.Partition))
	}
	return ""
}

type franzConsumerAttrsExtractor struct{}

func (e *franzConsumerAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request franzKafkaConsumerRequest) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (e *franzConsumerAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request franzKafkaConsumerRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, ctx
}

// --- Builder functions ---

func buildFranzKafkaProducerInstrumenter() instrumenter.Instrumenter[franzKafkaProducerRequest, any] {
	builder := instrumenter.Builder[franzKafkaProducerRequest, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FRANZ_GO_PRODUCER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[franzKafkaProducerRequest, any]{
			Getter:        franzProducerAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[franzKafkaProducerRequest]{}).
		SetSpanStatusExtractor(&franzProducerStatusExtractor{}).
		AddAttributesExtractor(&franzProducerAttrsExtractor{}).
		BuildPropagatingToDownstreamInstrumenter(
			func(req franzKafkaProducerRequest) propagation.TextMapCarrier {
				if req.Record == nil {
					return nil
				}
				return NewRecordCarrier(req.Record)
			},
			otel.GetTextMapPropagator(),
		)
}

func buildFranzKafkaConsumerInstrumenter() instrumenter.Instrumenter[franzKafkaConsumerRequest, any] {
	builder := instrumenter.Builder[franzKafkaConsumerRequest, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FRANZ_GO_CONSUMER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[franzKafkaConsumerRequest, any]{
			Getter:        franzConsumerAttrsGetter{},
			OperationName: message.PROCESS,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[franzKafkaConsumerRequest]{}).
		SetSpanStatusExtractor(&franzConsumerStatusExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[franzKafkaConsumerRequest, any, franzConsumerAttrsGetter]{
			Operation: message.PROCESS,
		}).
		AddAttributesExtractor(&franzConsumerAttrsExtractor{}).
		BuildPropagatingFromUpstreamInstrumenter(
			func(req franzKafkaConsumerRequest) propagation.TextMapCarrier {
				if req.Record == nil {
					return nil
				}
				return NewRecordCarrier(req.Record)
			},
			otel.GetTextMapPropagator(),
		)
}
