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

package eino

import (
	"context"
	"fmt"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

var _ ai.CommonAttrsGetter[einoLLMRequest, any] = einoLLMAttrsGetter{}

var _ ai.LLMAttrsGetter[einoLLMRequest, einoLLMResponse] = einoLLMAttrsGetter{}

type einoLLMAttrsGetter struct{}

func (e einoLLMAttrsGetter) GetAIOperationName(request einoLLMRequest) string {
	return request.operationName
}

func (e einoLLMAttrsGetter) GetAISystem(request einoLLMRequest) string {
	return "eino"
}

func (e einoLLMAttrsGetter) GetAIRequestModel(request einoLLMRequest) string {
	return request.modelName
}

func (e einoLLMAttrsGetter) GetAIRequestEncodingFormats(request einoLLMRequest) []string {
	return request.encodingFormats
}

var _ instrumenter.AttributesExtractor[einoLLMRequest, einoLLMResponse] = LLMExperimentalAttributeExtractor{}

type LLMExperimentalAttributeExtractor struct {
	Base      ai.AICommonAttrsExtractor[einoLLMRequest, einoLLMResponse, einoLLMAttrsGetter]
	LLMGetter einoLLMAttrsGetter
}

func (l LLMExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request einoLLMRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = l.Base.OnStart(attributes, parentContext, request)

	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIRequestModelKey,
		Value: attribute.StringValue(l.LLMGetter.GetAIRequestModel(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestEncodingFormatsKey,
		Value: attribute.StringSliceValue(l.LLMGetter.GetAIRequestEncodingFormats(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestMaxTokensKey,
		Value: attribute.Int64Value(l.LLMGetter.GetAIRequestMaxTokens(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestFrequencyPenaltyKey,
		Value: attribute.Float64Value(l.LLMGetter.GetAIRequestFrequencyPenalty(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestPresencePenaltyKey,
		Value: attribute.Float64Value(l.LLMGetter.GetAIRequestPresencePenalty(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestStopSequencesKey,
		Value: attribute.StringSliceValue(l.LLMGetter.GetAIRequestStopSequences(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTemperatureKey,
		Value: attribute.Float64Value(l.LLMGetter.GetAIRequestTemperature(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTopKKey,
		Value: attribute.Float64Value(l.LLMGetter.GetAIRequestTopK(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestTopPKey,
		Value: attribute.Float64Value(l.LLMGetter.GetAIRequestTopP(request)),
	}, attribute.KeyValue{
		Key:   semconv.ServerAddressKey,
		Value: attribute.StringValue(l.LLMGetter.GetAIServerAddress(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIRequestSeedKey,
		Value: attribute.Int64Value(l.LLMGetter.GetAIRequestSeed(request)),
	})
	for i, in := range request.input {
		if in != nil && len(in.Content) > 0 {
			attributes = append(attributes, attribute.String(fmt.Sprintf("gen_ai.prompt.%d.role", i), string(in.Role)))
			attributes = append(attributes, attribute.String(fmt.Sprintf("gen_ai.prompt.%d.content", i), in.Content))
		}
	}
	if l.Base.AttributesFilter != nil {
		attributes = l.Base.AttributesFilter(attributes)
	}
	return attributes, parentContext
}

func (l LLMExperimentalAttributeExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request einoLLMRequest, response einoLLMResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = l.Base.OnEnd(attributes, ctx, request, response, err)

	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.GenAIResponseFinishReasonsKey,
		Value: attribute.StringSliceValue(l.LLMGetter.GetAIResponseFinishReasons(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIResponseIDKey,
		Value: attribute.StringValue(l.LLMGetter.GetAIResponseID(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIResponseModelKey,
		Value: attribute.StringValue(l.LLMGetter.GetAIResponseModel(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIUsageInputTokensKey,
		Value: attribute.Int64Value(l.LLMGetter.GetAIUsageInputTokens(request)),
	}, attribute.KeyValue{
		Key:   semconv.GenAIUsageOutputTokensKey,
		Value: attribute.Int64Value(l.LLMGetter.GetAIUsageOutputTokens(request, response)),
	}, attribute.String("gen_ai.completion.0.content", response.output),
		attribute.Int64("gen_ai.usage.total_tokens", response.usageTotalTokens))

	return attributes, ctx
}

func (e einoLLMAttrsGetter) GetAIRequestFrequencyPenalty(request einoLLMRequest) float64 {
	return request.frequencyPenalty
}

func (e einoLLMAttrsGetter) GetAIRequestPresencePenalty(request einoLLMRequest) float64 {
	return request.presencePenalty
}

func (e einoLLMAttrsGetter) GetAIResponseFinishReasons(request einoLLMRequest, response einoLLMResponse) []string {
	return response.responseFinishReasons
}

func (e einoLLMAttrsGetter) GetAIResponseModel(request einoLLMRequest, response einoLLMResponse) string {
	return response.responseModel
}

func (e einoLLMAttrsGetter) GetAIRequestMaxTokens(request einoLLMRequest) int64 {
	return request.maxTokens
}

func (e einoLLMAttrsGetter) GetAIUsageInputTokens(request einoLLMRequest) int64 {
	return request.usageInputTokens
}

func (e einoLLMAttrsGetter) GetAIUsageOutputTokens(request einoLLMRequest, response einoLLMResponse) int64 {
	return response.usageOutputTokens
}

func (e einoLLMAttrsGetter) GetAIRequestStopSequences(request einoLLMRequest) []string {
	return request.stopSequences
}

func (e einoLLMAttrsGetter) GetAIRequestTemperature(request einoLLMRequest) float64 {
	return request.temperature
}

func (e einoLLMAttrsGetter) GetAIRequestTopK(request einoLLMRequest) float64 {
	return request.topK
}

func (e einoLLMAttrsGetter) GetAIRequestTopP(request einoLLMRequest) float64 {
	return request.topP
}

func (e einoLLMAttrsGetter) GetAIResponseID(request einoLLMRequest, response einoLLMResponse) string {
	return response.responseID
}

func (e einoLLMAttrsGetter) GetAIServerAddress(request einoLLMRequest) string {
	return request.serverAddress
}

func (e einoLLMAttrsGetter) GetAIRequestSeed(request einoLLMRequest) int64 {
	return request.seed
}

func BuildEinoLLMInstrumenter() instrumenter.Instrumenter[einoLLMRequest, einoLLMResponse] {
	builder := instrumenter.Builder[einoLLMRequest, einoLLMResponse]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[einoLLMRequest, einoLLMResponse]{Getter: einoLLMAttrsGetter{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[einoLLMRequest]{}).
		AddAttributesExtractor(&LLMExperimentalAttributeExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.EINO_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(ai.AIClientMetrics("eino-llm")).
		BuildInstrumenter()
}
