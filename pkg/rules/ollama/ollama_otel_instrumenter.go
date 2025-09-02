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
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
)

const (
	OLLAMA_SCOPE_NAME = "github.com/alibaba/loongsuite-go-agent/pkg/rules/ollama"
)

type ollamaAttrsGetter struct{}

func (o ollamaAttrsGetter) GetAISystem(request ollamaRequest) string {
	return "ollama"
}

func (o ollamaAttrsGetter) GetAIRequestModel(request ollamaRequest) string {
	return request.model
}

func (o ollamaAttrsGetter) GetAIRequestTemperature(request ollamaRequest) float64 {
	// Temperature parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestMaxTokens(request ollamaRequest) int64 {
	// Max tokens parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestTopP(request ollamaRequest) float64 {
	// TopP parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestTopK(request ollamaRequest) float64 {
	// TopK parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestStopSequences(request ollamaRequest) []string {
	// Stop sequences not captured in this implementation
	return nil
}

func (o ollamaAttrsGetter) GetAIRequestFrequencyPenalty(request ollamaRequest) float64 {
	// Frequency penalty parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestPresencePenalty(request ollamaRequest) float64 {
	// Presence penalty parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIRequestIsStream(request ollamaRequest) bool {
	// Return true if this is a streaming request
	return request.isStreaming
}

func (o ollamaAttrsGetter) GetAIOperationName(request ollamaRequest) string {
	return request.operationType
}

func (o ollamaAttrsGetter) GetAIRequestEncodingFormats(request ollamaRequest) []string {
	// Encoding formats not captured in this implementation
	return nil
}

func (o ollamaAttrsGetter) GetAIRequestSeed(request ollamaRequest) int64 {
	// Seed parameter not captured in this implementation
	return 0
}

func (o ollamaAttrsGetter) GetAIResponseModel(request ollamaRequest, response ollamaResponse) string {
	// Model comes from request
	return request.model
}

func (o ollamaAttrsGetter) GetAIUsageInputTokens(request ollamaRequest) int64 {
	return int64(request.promptTokens)
}

func (o ollamaAttrsGetter) GetAIUsageOutputTokens(request ollamaRequest, response ollamaResponse) int64 {
	return int64(request.completionTokens)
}

func (o ollamaAttrsGetter) GetStreamingMetrics(response ollamaResponse) map[string]interface{} {
	metrics := make(map[string]interface{})

	if response.streamingMetrics != nil {
		metrics["gen_ai.response.streaming"] = true
		metrics["gen_ai.response.ttft_ms"] = response.streamingMetrics.getTTFTMillis()
		metrics["gen_ai.response.chunk_count"] = response.streamingMetrics.chunkCount

		if response.streamingMetrics.tokenRate > 0 {
			metrics["gen_ai.response.tokens_per_second"] = response.streamingMetrics.tokenRate
		}

		if response.streamingMetrics.endTime != nil {
			streamDuration := response.streamingMetrics.endTime.Sub(response.streamingMetrics.startTime).Milliseconds()
			metrics["gen_ai.response.stream_duration_ms"] = streamDuration
		}
	} else {
		metrics["gen_ai.response.streaming"] = false
	}

	return metrics
}

func (o ollamaAttrsGetter) GetAIResponseFinishReasons(request ollamaRequest, response ollamaResponse) []string {
	if response.err != nil {
		return []string{"error"}
	}
	return []string{"stop"}
}

func (o ollamaAttrsGetter) GetAIResponseID(request ollamaRequest, response ollamaResponse) string {
	// Response ID not available in Ollama API
	return ""
}

func (o ollamaAttrsGetter) GetAIServerAddress(request ollamaRequest) string {
	// Server address not captured in this implementation
	return ""
}

func BuildOllamaLLMInstrumenter() instrumenter.Instrumenter[ollamaRequest, ollamaResponse] {
	builder := instrumenter.Builder[ollamaRequest, ollamaResponse]{}
	getter := ollamaAttrsGetter{}

	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[ollamaRequest, ollamaResponse]{Getter: getter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[ollamaRequest]{}).
		AddAttributesExtractor(&ai.AILLMAttrsExtractor[ollamaRequest, ollamaResponse, ollamaAttrsGetter, ollamaAttrsGetter]{}).
		AddAttributesExtractor(&streamingAttributesExtractor{}).
		AddAttributesExtractor(&costAttributesExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    OLLAMA_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

type streamingAttributesExtractor struct{}

func (s *streamingAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request ollamaRequest) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (s *streamingAttributesExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request ollamaRequest, response ollamaResponse, err error) ([]attribute.KeyValue, context.Context) {
	if response.streamingMetrics != nil {
		attributes = append(attributes, attribute.Bool("gen_ai.response.streaming", true))
		if ttft := response.streamingMetrics.getTTFTMillis(); ttft > 0 {
			attributes = append(attributes, attribute.Int64("gen_ai.response.ttft_ms", ttft))
		}
		attributes = append(attributes, attribute.Int("gen_ai.response.chunk_count", response.streamingMetrics.chunkCount))
		if response.streamingMetrics.tokenRate > 0 {
			attributes = append(attributes, attribute.Float64("gen_ai.response.tokens_per_second", response.streamingMetrics.tokenRate))
		}
		if response.streamingMetrics.endTime != nil {
			streamDuration := response.streamingMetrics.endTime.Sub(response.streamingMetrics.startTime).Milliseconds()
			attributes = append(attributes, attribute.Int64("gen_ai.response.stream_duration_ms", streamDuration))
		}
	} else if request.isStreaming {
		attributes = append(attributes, attribute.Bool("gen_ai.response.streaming", true))
		attributes = append(attributes, attribute.Bool("gen_ai.response.partial", true))
	} else {
		attributes = append(attributes, attribute.Bool("gen_ai.response.streaming", false))
	}

	return attributes, context
}

type costAttributesExtractor struct{}

func (c *costAttributesExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request ollamaRequest) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (c *costAttributesExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request ollamaRequest, response ollamaResponse, err error) ([]attribute.KeyValue, context.Context) {
	if response.costMetrics != nil {
		attributes = append(attributes,
			attribute.Float64("gen_ai.cost.input_tokens_usd", response.costMetrics.InputCost),
			attribute.Float64("gen_ai.cost.output_tokens_usd", response.costMetrics.OutputCost),
			attribute.Float64("gen_ai.cost.total_usd", response.costMetrics.TotalCost),
			attribute.String("gen_ai.cost.currency", string(response.costMetrics.Currency)),
			attribute.String("gen_ai.cost.model_pricing_tier", response.costMetrics.PricingTier),
		)
		
		budgetTracker := globalBudget
		if budgetTracker != nil {
			status, percentage, remaining := budgetTracker.GetStatus()
			attributes = append(attributes, attribute.String("gen_ai.budget.status", string(status)))
			attributes = append(attributes, attribute.Float64("gen_ai.budget.usage_percentage", percentage))
			attributes = append(attributes, attribute.Float64("gen_ai.budget.remaining_usd", remaining))
			thresholdExceeded := percentage >= 100
			attributes = append(attributes, attribute.Bool("gen_ai.budget.threshold_exceeded", thresholdExceeded))
		}
		if response.costMetrics.EstimatedInput {
			attributes = append(attributes, attribute.Bool("gen_ai.cost.input_estimated", true))
		}
	}
	
	return attributes, context
}

var ollamaInstrumenter = BuildOllamaLLMInstrumenter()
