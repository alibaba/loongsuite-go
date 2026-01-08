// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ollama

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
)

type ollamaAttrsGetter struct{}

func (o ollamaAttrsGetter) GetAISystem(request ollamaRequest) string {
	return "ollama"
}

func (o ollamaAttrsGetter) GetAIRequestModel(request ollamaRequest) string {
	return request.model
}

func (o ollamaAttrsGetter) GetAIRequestTemperature(request ollamaRequest) float64 {
	return request.temperature
}

func (o ollamaAttrsGetter) GetAIRequestMaxTokens(request ollamaRequest) int64 {
	return request.maxTokens
}

func (o ollamaAttrsGetter) GetAIRequestTopP(request ollamaRequest) float64 {
	return request.topP
}

func (o ollamaAttrsGetter) GetAIRequestTopK(request ollamaRequest) float64 {
	return request.topK
}

func (o ollamaAttrsGetter) GetAIRequestStopSequences(request ollamaRequest) []string {
	return request.stopSequences
}

func (o ollamaAttrsGetter) GetAIRequestFrequencyPenalty(request ollamaRequest) float64 {
	return request.frequencyPenalty
}

func (o ollamaAttrsGetter) GetAIRequestPresencePenalty(request ollamaRequest) float64 {
	return request.presencePenalty
}

func (o ollamaAttrsGetter) GetAIRequestIsStream(request ollamaRequest) bool {
	return request.isStreaming
}

func (o ollamaAttrsGetter) GetAIOperationName(request ollamaRequest) string {
	if request.modelOperation != "" {
		return request.modelOperation
	}
	return request.operationType
}

func (o ollamaAttrsGetter) GetAIRequestEncodingFormats(request ollamaRequest) []string {
	return nil
}

func (o ollamaAttrsGetter) GetAIRequestSeed(request ollamaRequest) int64 {
	return request.seed
}

func (o ollamaAttrsGetter) GetAIInput(request ollamaRequest) string {
	return request.input
}

func (o ollamaAttrsGetter) GetAIOutput(response ollamaResponse) string { // Changed from response.output to response.output
	return response.content
}

func (o ollamaAttrsGetter) GetAIResponseModel(request ollamaRequest, response ollamaResponse) string {
	return request.model
}

func (o ollamaAttrsGetter) GetAIUsageInputTokens(request ollamaRequest) int64 {
	return int64(request.promptTokens)
}

func (o ollamaAttrsGetter) GetAIUsageOutputTokens(request ollamaRequest, response ollamaResponse) int64 {
	return int64(request.completionTokens)
}

func (o ollamaAttrsGetter) GetStreamingMetrics(response ollamaResponse) map[string]interface{} {
	return make(map[string]interface{})
}

func (o ollamaAttrsGetter) GetAIResponseFinishReasons(request ollamaRequest, response ollamaResponse) []string {
	if response.err != nil {
		return []string{"error"}
	}
	return []string{"stop"}
}

func (o ollamaAttrsGetter) GetAIResponseID(request ollamaRequest, response ollamaResponse) string {
	return ""
}

func (o ollamaAttrsGetter) GetAIServerAddress(request ollamaRequest) string {
	return request.serverAddress
}

func BuildOllamaLLMInstrumenter() instrumenter.Instrumenter[ollamaRequest, ollamaResponse] {
	builder := instrumenter.Builder[ollamaRequest, ollamaResponse]{}
	getter := ollamaAttrsGetter{}

	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[ollamaRequest, ollamaResponse]{Getter: getter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[ollamaRequest]{}).
		AddAttributesExtractor(&ai.AILLMAttrsExtractor[ollamaRequest, ollamaResponse, ollamaAttrsGetter, ollamaAttrsGetter]{}).
		AddAttributesExtractor(&embeddingAttributesExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.OLLAMA_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(ai.AIClientMetrics("ollama")).
		BuildInstrumenter()
}

type embeddingAttributesExtractor struct{}

func (e *embeddingAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request ollamaRequest) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (e *embeddingAttributesExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request ollamaRequest, response ollamaResponse, err error) ([]attribute.KeyValue, context.Context) {
	if request.operationType == "embed" || request.operationType == "embeddings" {
		attributes = append(attributes,
			attribute.Int("gen_ai.embedding.count", request.embeddingCount),
			attribute.Int("gen_ai.embedding.dimensions", request.embeddingDim),
		)
	}
	return attributes, context
}

var ollamaInstrumenter = BuildOllamaLLMInstrumenter()
