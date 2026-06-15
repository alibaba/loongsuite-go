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

package openai

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// openaiCommonRequest implements the common attributes getter
type openaiCommonRequest struct{}

func (openaiCommonRequest) GetAIOperationName(request openaiRequest) string {
	return request.operationName
}

func (openaiCommonRequest) GetAISystem(request openaiRequest) string {
	return "openai"
}

// openaiLLMRequest implements the LLM attributes getter
type openaiLLMRequest struct{}

func (openaiLLMRequest) GetAIRequestModel(request openaiRequest) string {
	return request.modelName
}

func (openaiLLMRequest) GetAIRequestEncodingFormats(request openaiRequest) []string {
	return nil
}

func (openaiLLMRequest) GetAIRequestFrequencyPenalty(request openaiRequest) float64 {
	return request.frequencyPenalty
}

func (openaiLLMRequest) GetAIRequestPresencePenalty(request openaiRequest) float64 {
	return request.presencePenalty
}

func (openaiLLMRequest) GetAIResponseFinishReasons(request openaiRequest, response openaiResponse) []string {
	return response.finishReasons
}

func (openaiLLMRequest) GetAIResponseModel(request openaiRequest, response openaiResponse) string {
	return response.responseModel
}

func (openaiLLMRequest) GetAIRequestMaxTokens(request openaiRequest) int64 {
	return request.maxTokens
}

func (openaiLLMRequest) GetAIUsageInputTokens(request openaiRequest) int64 {
	return request.inputTokens
}

func (openaiLLMRequest) GetAIUsageOutputTokens(request openaiRequest, response openaiResponse) int64 {
	return response.usageOutputTokens
}

func (openaiLLMRequest) GetAIRequestStopSequences(request openaiRequest) []string {
	return request.stopSequences
}

func (openaiLLMRequest) GetAIRequestTemperature(request openaiRequest) float64 {
	return request.temperature
}

func (openaiLLMRequest) GetAIRequestTopK(request openaiRequest) float64 {
	return 0
}

func (openaiLLMRequest) GetAIRequestTopP(request openaiRequest) float64 {
	return request.topP
}

func (openaiLLMRequest) GetAIResponseID(request openaiRequest, response openaiResponse) string {
	return response.responseID
}

func (openaiLLMRequest) GetAIServerAddress(request openaiRequest) string {
	return request.serverAddress
}

func (openaiLLMRequest) GetAIRequestSeed(request openaiRequest) int64 {
	return request.seed
}

func (openaiLLMRequest) GetAIInput(request openaiRequest) string {
	return request.inputMessages
}

func (openaiLLMRequest) GetAIOutput(response openaiResponse) string {
	return response.outputMessages
}

// OpenAIExperimentalAttributeExtractor adds OpenAI-specific experimental attributes
type OpenAIExperimentalAttributeExtractor struct {
	Base ai.AILLMAttrsExtractor[openaiRequest, openaiResponse, openaiCommonRequest, openaiLLMRequest]
}

func (o OpenAIExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request openaiRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = o.Base.OnStart(attributes, parentContext, request)
	return attributes, parentContext
}

func (o OpenAIExperimentalAttributeExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request openaiRequest, response openaiResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = o.Base.OnEnd(attributes, ctx, request, response, err)
	// Add total tokens if available
	if response.usageTotalTokens > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   "gen_ai.usage.total_tokens",
			Value: attribute.Int64Value(response.usageTotalTokens),
		})
	}

	return attributes, ctx
}

// BuildOpenAIClientOtelInstrumenter builds the OpenAI client instrumenter
func BuildOpenAIClientOtelInstrumenter() instrumenter.Instrumenter[openaiRequest, openaiResponse] {
	builder := instrumenter.Builder[openaiRequest, openaiResponse]{}

	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[openaiRequest, openaiResponse]{
			Getter: openaiCommonRequest{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[openaiRequest]{}).
		AddAttributesExtractor(&OpenAIExperimentalAttributeExtractor{
			Base: ai.AILLMAttrsExtractor[openaiRequest, openaiResponse, openaiCommonRequest, openaiLLMRequest]{
				Base: ai.AICommonAttrsExtractor[openaiRequest, openaiResponse, openaiCommonRequest]{
					CommonGetter: openaiCommonRequest{},
				},
				LLMGetter: openaiLLMRequest{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    "loongsuite.instrumentation.openai",
			Version: version.Tag,
		}).
		AddOperationListeners(ai.AIClientMetrics("openai-client")).
		BuildInstrumenter()
}

// AIMetricsRecorder records AI-specific metrics
type AIMetricsRecorder struct {
	instrumenter instrumenter.Instrumenter[openaiRequest, openaiResponse]
}

func NewAIMetricsRecorder() *AIMetricsRecorder {
	return &AIMetricsRecorder{
		instrumenter: BuildOpenAIClientOtelInstrumenter(),
	}
}

func (a *AIMetricsRecorder) Start(ctx context.Context, request openaiRequest) context.Context {
	return a.instrumenter.Start(ctx, request)
}

func (a *AIMetricsRecorder) End(ctx context.Context, request openaiRequest, response openaiResponse, err error) {
	a.instrumenter.End(ctx, request, response, err)
}
