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

	// Always add model
	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIRequestModelKey,
		Value: attribute.StringValue(h.LLMGetter.GetAIRequestModel(request)),
	})

	// Only add optional parameters if they have non-zero values
	if maxTokens := h.LLMGetter.GetAIRequestMaxTokens(request); maxTokens > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestMaxTokensKey,
			Value: attribute.Int64Value(maxTokens),
		})
	}
	if frequencyPenalty := h.LLMGetter.GetAIRequestFrequencyPenalty(request); frequencyPenalty != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestFrequencyPenaltyKey,
			Value: attribute.Float64Value(frequencyPenalty),
		})
	}
	if presencePenalty := h.LLMGetter.GetAIRequestPresencePenalty(request); presencePenalty != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestPresencePenaltyKey,
			Value: attribute.Float64Value(presencePenalty),
		})
	}
	if stopSequences := h.LLMGetter.GetAIRequestStopSequences(request); len(stopSequences) > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestStopSequencesKey,
			Value: attribute.StringSliceValue(stopSequences),
		})
	}
	if temperature := h.LLMGetter.GetAIRequestTemperature(request); temperature != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestTemperatureKey,
			Value: attribute.Float64Value(temperature),
		})
	}
	if topK := h.LLMGetter.GetAIRequestTopK(request); topK != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestTopKKey,
			Value: attribute.Float64Value(topK),
		})
	}
	if topP := h.LLMGetter.GetAIRequestTopP(request); topP != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestTopPKey,
			Value: attribute.Float64Value(topP),
		})
	}
	if seed := h.LLMGetter.GetAIRequestSeed(request); seed != 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIRequestSeedKey,
			Value: attribute.Int64Value(seed),
		})
	}
	if input := h.LLMGetter.GetAIInput(request); input != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv7.GenAIInputMessagesKey,
			Value: attribute.StringValue(input),
		})
	}

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

	// Only add attributes with non-zero/non-empty values
	if finishReasons := h.LLMGetter.GetAIResponseFinishReasons(request, response); len(finishReasons) > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIResponseFinishReasonsKey,
			Value: attribute.StringSliceValue(finishReasons),
		})
	}
	if responseModel := h.LLMGetter.GetAIResponseModel(request, response); responseModel != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIResponseModelKey,
			Value: attribute.StringValue(responseModel),
		})
	}
	if inputTokens := h.LLMGetter.GetAIUsageInputTokens(request); inputTokens > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIUsageInputTokensKey,
			Value: attribute.Int64Value(inputTokens),
		})
	}
	if outputTokens := h.LLMGetter.GetAIUsageOutputTokens(request, response); outputTokens > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIUsageOutputTokensKey,
			Value: attribute.Int64Value(outputTokens),
		})
	}
	if output := h.LLMGetter.GetAIOutput(response); output != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv7.GenAIOutputMessagesKey,
			Value: attribute.StringValue(output),
		})
	}

	// Only add response id if it's not empty
	if responseID := h.LLMGetter.GetAIResponseID(request, response); responseID != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.GenAIResponseIDKey,
			Value: attribute.StringValue(responseID),
		})
	}

	return attributes, context
}
