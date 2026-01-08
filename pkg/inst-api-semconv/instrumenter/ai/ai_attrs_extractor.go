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

package ai

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	semconv7 "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// TODO: remove server.address and put it into NetworkAttributesExtractor

type AICommonAttrsExtractor[REQUEST any, RESPONSE any, GETTER1 CommonAttrsGetter[REQUEST, RESPONSE]] struct {
	CommonGetter     GETTER1
	AttributesFilter func(attrs []attribute.KeyValue) []attribute.KeyValue
}

func (h *AICommonAttrsExtractor[REQUEST, RESPONSE, GETTER1]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIOperationNameKey,
		Value: attribute.StringValue(h.CommonGetter.GetAIOperationName(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAISystemKey,
		Value: attribute.StringValue(h.CommonGetter.GetAISystem(request)),
	})
	return attributes, parentContext
}

func (h *AICommonAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	if err != nil {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.ErrorTypeKey,
			Value: attribute.StringValue(err.Error()),
		})
	}
	return attributes, context
}

type AILLMAttrsExtractor[REQUEST any, RESPONSE any, GETTER1 CommonAttrsGetter[REQUEST, RESPONSE], GETTER2 LLMAttrsGetter[REQUEST, RESPONSE]] struct {
	Base      AICommonAttrsExtractor[REQUEST, RESPONSE, GETTER1]
	LLMGetter GETTER2
}

func (h *AILLMAttrsExtractor[REQUEST, RESPONSE, GETTER1, GETTER2]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = h.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIRequestModelKey,
		Value: attribute.StringValue(h.LLMGetter.GetAIRequestModel(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestMaxTokensKey,
		Value: attribute.Int64Value(h.LLMGetter.GetAIRequestMaxTokens(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestFrequencyPenaltyKey,
		Value: attribute.Float64Value(h.LLMGetter.GetAIRequestFrequencyPenalty(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestPresencePenaltyKey,
		Value: attribute.Float64Value(h.LLMGetter.GetAIRequestPresencePenalty(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestStopSequencesKey,
		Value: attribute.StringSliceValue(h.LLMGetter.GetAIRequestStopSequences(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTemperatureKey,
		Value: attribute.Float64Value(h.LLMGetter.GetAIRequestTemperature(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTopKKey,
		Value: attribute.Float64Value(h.LLMGetter.GetAIRequestTopK(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTopPKey,
		Value: attribute.Float64Value(h.LLMGetter.GetAIRequestTopP(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestSeedKey,
		Value: attribute.Int64Value(h.LLMGetter.GetAIRequestSeed(request)),
	}, attribute.KeyValue{
		Key:   semconv7.GenAIInputMessagesKey,
		Value: attribute.StringValue(h.LLMGetter.GetAIInput(request)),
	})

	if serverAddress := h.LLMGetter.GetAIServerAddress(request); serverAddress != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.ServerAddressKey,
			Value: attribute.StringValue(serverAddress),
		})
	}

	if h.Base.AttributesFilter != nil {
		attributes = h.Base.AttributesFilter(attributes)
	}
	return attributes, parentContext
}
func (h *AILLMAttrsExtractor[REQUEST, RESPONSE, GETTER1, GETTER2]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	attributes, context = h.Base.OnEnd(attributes, context, request, response, err)

	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIResponseFinishReasonsKey,
		Value: attribute.StringSliceValue(h.LLMGetter.GetAIResponseFinishReasons(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIResponseModelKey,
		Value: attribute.StringValue(h.LLMGetter.GetAIResponseModel(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIUsageInputTokensKey,
		Value: attribute.Int64Value(h.LLMGetter.GetAIUsageInputTokens(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIUsageOutputTokensKey,
		Value: attribute.Int64Value(h.LLMGetter.GetAIUsageOutputTokens(request, response)),
	}, attribute.KeyValue{
		Key:   semconv7.GenAIOutputMessagesKey,
		Value: attribute.StringValue(h.LLMGetter.GetAIOutput(response)),
	})

	// Only add response id if it's not empty
	if responseID := h.LLMGetter.GetAIResponseID(request, response); responseID != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIResponseIDKey,
			Value: attribute.StringValue(responseID),
		})
	}

	return attributes, context
}
