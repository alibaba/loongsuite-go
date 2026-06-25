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

package utilgenai

import "go.opentelemetry.io/otel/attribute"

// ============================================================================
// Basic Semantic Conventions
//
// These attribute keys follow the OpenTelemetry semantic conventions for GenAI
// and are used for basic LLM instrumentation (chat, text_completion, generate_content).
// See: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/README.md
// ============================================================================

const (
	// GenAI operation attributes (Basic)
	AttrGenAIOperationName = "gen_ai.operation.name"
	AttrGenAIProviderName  = "gen_ai.provider.name"

	// GenAI request attributes (Basic)
	AttrGenAIRequestModel            = "gen_ai.request.model"
	AttrGenAIRequestTemperature      = "gen_ai.request.temperature"
	AttrGenAIRequestTopP             = "gen_ai.request.top_p"
	AttrGenAIRequestFrequencyPenalty = "gen_ai.request.frequency_penalty"
	AttrGenAIRequestPresencePenalty  = "gen_ai.request.presence_penalty"
	AttrGenAIRequestMaxTokens        = "gen_ai.request.max_tokens"
	AttrGenAIRequestStopSequences    = "gen_ai.request.stop_sequences"
	AttrGenAIRequestSeed             = "gen_ai.request.seed"

	// GenAI response attributes (Basic)
	AttrGenAIResponseModel         = "gen_ai.response.model"
	AttrGenAIResponseID            = "gen_ai.response.id"
	AttrGenAIResponseFinishReasons = "gen_ai.response.finish_reasons"

	// GenAI usage attributes (Basic)
	AttrGenAIUsageInputTokens  = "gen_ai.usage.input_tokens"
	AttrGenAIUsageOutputTokens = "gen_ai.usage.output_tokens"
	AttrGenAIUsageTotalTokens  = "gen_ai.usage.total_tokens"

	// GenAI message content attributes - experimental (Basic)
	AttrGenAIInputMessages      = "gen_ai.input.messages"
	AttrGenAIOutputMessages     = "gen_ai.output.messages"
	AttrGenAISystemInstructions = "gen_ai.system_instructions"
	AttrGenAIToolDefinitions    = "gen_ai.tool_definitions"

	// GenAI token type attribute for metrics (Basic)
	AttrGenAITokenType = "gen_ai.token.type"

	// Error attributes (Basic)
	AttrErrorType = "error.type"

	// ========================================================================
	// Extended Semantic Conventions (LoongSuite Extension)
	//
	// These attribute keys are extensions for additional GenAI operations
	// beyond basic LLM calls.
	// ========================================================================

	// Span kind attribute (LoongSuite Extension)
	AttrGenAISpanKind = "gen_ai.span.kind"

	// Agent attributes (LoongSuite Extension)
	AttrGenAIAgentName = "gen_ai.agent.name"
	AttrGenAIAgentID   = "gen_ai.agent.id"

	// Tool attributes (LoongSuite Extension)
	AttrGenAIToolName   = "gen_ai.tool.name"
	AttrGenAIToolCallID = "gen_ai.tool.call.id"
	AttrGenAIToolInput  = "gen_ai.tool.input"
	AttrGenAIToolOutput = "gen_ai.tool.output"

	// Embedding attributes (LoongSuite Extension)
	AttrGenAIEmbeddingInputCount = "gen_ai.embedding.input_count"
	AttrGenAIEmbeddingDimensions = "gen_ai.embedding.dimensions"

	// Retrieve attributes (LoongSuite Extension)
	AttrGenAIRetrieveQuery          = "gen_ai.retrieve.query"
	AttrGenAIRetrieveTopK           = "gen_ai.retrieve.top_k"
	AttrGenAIRetrieveDocumentCount  = "gen_ai.retrieve.document_count"
	AttrGenAIRetrieveDataSourceName = "gen_ai.retrieve.data_source_name"

	// Rerank attributes (LoongSuite Extension)
	AttrGenAIRerankQuery       = "gen_ai.rerank.query"
	AttrGenAIRerankModel       = "gen_ai.rerank.model"
	AttrGenAIRerankTopN        = "gen_ai.rerank.top_n"
	AttrGenAIRerankInputCount  = "gen_ai.rerank.input_count"
	AttrGenAIRerankOutputCount = "gen_ai.rerank.output_count"
)

// SpanKind values for GenAI operations (LoongSuite Extension)
type SpanKindValue string

const (
	// Basic span kind
	SpanKindLLM SpanKindValue = "llm"

	// Extended span kinds (LoongSuite Extension)
	SpanKindEmbedding SpanKindValue = "embedding"
	SpanKindAgent     SpanKindValue = "agent"
	SpanKindTool      SpanKindValue = "tool"
	SpanKindRetriever SpanKindValue = "retriever"
	SpanKindReranker  SpanKindValue = "reranker"
)

// TokenType values for metrics
type TokenType string

const (
	TokenTypeInput  TokenType = "input"
	TokenTypeOutput TokenType = "output"
)

// Metric names for GenAI (Basic)
const (
	MetricGenAIClientOperationDuration = "gen_ai.client.operation.duration"
	MetricGenAIClientTokenUsage        = "gen_ai.client.token.usage"
)

// ============================================================================
// Basic Attribute Helper Functions
// ============================================================================

// GenAIOperationName creates an attribute for the operation name.
func GenAIOperationName(name OperationName) attribute.KeyValue {
	return attribute.String(AttrGenAIOperationName, string(name))
}

// GenAIProviderName creates an attribute for the provider name.
func GenAIProviderName(name string) attribute.KeyValue {
	return attribute.String(AttrGenAIProviderName, name)
}

// GenAIRequestModel creates an attribute for the request model.
func GenAIRequestModel(model string) attribute.KeyValue {
	return attribute.String(AttrGenAIRequestModel, model)
}

// GenAIResponseModel creates an attribute for the response model.
func GenAIResponseModel(model string) attribute.KeyValue {
	return attribute.String(AttrGenAIResponseModel, model)
}

// GenAIResponseID creates an attribute for the response ID.
func GenAIResponseID(id string) attribute.KeyValue {
	return attribute.String(AttrGenAIResponseID, id)
}

// GenAIResponseFinishReasons creates an attribute for finish reasons.
func GenAIResponseFinishReasons(reasons []string) attribute.KeyValue {
	return attribute.StringSlice(AttrGenAIResponseFinishReasons, reasons)
}

// GenAIUsageInputTokens creates an attribute for input token count.
func GenAIUsageInputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageInputTokens, count)
}

// GenAIUsageOutputTokens creates an attribute for output token count.
func GenAIUsageOutputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageOutputTokens, count)
}

// GenAIUsageTotalTokens creates an attribute for total token count.
func GenAIUsageTotalTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageTotalTokens, count)
}

// ============================================================================
// Extended Attribute Helper Functions (LoongSuite Extension)
// ============================================================================

// GenAISpanKind creates an attribute for the span kind.
// (LoongSuite Extension)
func GenAISpanKind(kind SpanKindValue) attribute.KeyValue {
	return attribute.String(AttrGenAISpanKind, string(kind))
}
