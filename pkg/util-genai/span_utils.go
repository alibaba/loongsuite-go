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

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ============================================================================
// Basic LLM Span Utilities
//
// These functions provide span attribute handling for LLM operations (chat,
// text_completion, generate_content), which is the core functionality for
// instrumenting language model calls.
// ============================================================================

// GetLLMSpanName returns the span name for an LLM invocation.
func GetLLMSpanName(invocation *LLMInvocation) string {
	if invocation.RequestModel != "" {
		return fmt.Sprintf("%s %s", invocation.OperationName, invocation.RequestModel)
	}
	return string(invocation.OperationName)
}

// GetLLMCommonAttributes returns common LLM attributes shared by all paths.
func GetLLMCommonAttributes(invocation *LLMInvocation) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		GenAIOperationName(invocation.OperationName),
		GenAISpanKind(SpanKindLLM),
	}

	if invocation.RequestModel != "" {
		attrs = append(attrs, GenAIRequestModel(invocation.RequestModel))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}

	return attrs
}

// GetLLMRequestAttributes returns GenAI request semantic convention attributes.
func GetLLMRequestAttributes(invocation *LLMInvocation) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	if invocation.Temperature != nil {
		attrs = append(attrs, attribute.Float64(AttrGenAIRequestTemperature, *invocation.Temperature))
	}
	if invocation.TopP != nil {
		attrs = append(attrs, attribute.Float64(AttrGenAIRequestTopP, *invocation.TopP))
	}
	if invocation.FrequencyPenalty != nil {
		attrs = append(attrs, attribute.Float64(AttrGenAIRequestFrequencyPenalty, *invocation.FrequencyPenalty))
	}
	if invocation.PresencePenalty != nil {
		attrs = append(attrs, attribute.Float64(AttrGenAIRequestPresencePenalty, *invocation.PresencePenalty))
	}
	if invocation.MaxTokens != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRequestMaxTokens, *invocation.MaxTokens))
	}
	if len(invocation.StopSequences) > 0 {
		attrs = append(attrs, attribute.StringSlice(AttrGenAIRequestStopSequences, invocation.StopSequences))
	}
	if invocation.Seed != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRequestSeed, *invocation.Seed))
	}

	return attrs
}

// GetLLMResponseAttributes returns GenAI response semantic convention attributes.
func GetLLMResponseAttributes(invocation *LLMInvocation) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	// Collect finish reasons
	var finishReasons []string
	if len(invocation.FinishReasons) > 0 {
		for _, reason := range invocation.FinishReasons {
			finishReasons = append(finishReasons, string(reason))
		}
	} else if len(invocation.OutputMessages) > 0 {
		for _, msg := range invocation.OutputMessages {
			if msg.FinishReason != "" {
				finishReasons = append(finishReasons, string(msg.FinishReason))
			}
		}
	}

	if len(finishReasons) > 0 {
		// De-duplicate finish reasons
		unique := make(map[string]bool)
		var uniqueReasons []string
		for _, reason := range finishReasons {
			if !unique[reason] {
				unique[reason] = true
				uniqueReasons = append(uniqueReasons, reason)
			}
		}
		if len(uniqueReasons) > 0 {
			attrs = append(attrs, GenAIResponseFinishReasons(uniqueReasons))
		}
	}

	if invocation.ResponseModelName != "" {
		attrs = append(attrs, GenAIResponseModel(invocation.ResponseModelName))
	}
	if invocation.ResponseID != "" {
		attrs = append(attrs, GenAIResponseID(invocation.ResponseID))
	}
	if invocation.InputTokens != nil {
		attrs = append(attrs, GenAIUsageInputTokens(*invocation.InputTokens))
	}
	if invocation.OutputTokens != nil {
		attrs = append(attrs, GenAIUsageOutputTokens(*invocation.OutputTokens))
	}

	// Calculate total tokens
	totalTokens := 0
	if invocation.InputTokens != nil {
		totalTokens += *invocation.InputTokens
	}
	if invocation.OutputTokens != nil {
		totalTokens += *invocation.OutputTokens
	}
	if totalTokens > 0 {
		attrs = append(attrs, GenAIUsageTotalTokens(totalTokens))
	}

	return attrs
}

// GetLLMMessageAttributesForSpan returns message attributes formatted for span (JSON string format).
// Returns empty slice if not in experimental mode or content capturing is disabled.
func GetLLMMessageAttributesForSpan(invocation *LLMInvocation) []attribute.KeyValue {
	if !IsExperimentalMode() || !ShouldCaptureContentInSpan() {
		return nil
	}

	var attrs []attribute.KeyValue

	if len(invocation.InputMessages) > 0 {
		attrs = append(attrs, attribute.String(AttrGenAIInputMessages, InputMessagesToJSON(invocation.InputMessages)))
	}
	if len(invocation.OutputMessages) > 0 {
		attrs = append(attrs, attribute.String(AttrGenAIOutputMessages, OutputMessagesToJSON(invocation.OutputMessages)))
	}
	if len(invocation.SystemInstruction) > 0 {
		attrs = append(attrs, attribute.String(AttrGenAISystemInstructions, SystemInstructionToJSON(invocation.SystemInstruction)))
	}
	if len(invocation.ToolDefinitions) > 0 {
		attrs = append(attrs, attribute.String(AttrGenAIToolDefinitions, ToolDefinitionsToJSON(invocation.ToolDefinitions)))
	}

	return attrs
}

// ApplyLLMFinishAttributes applies attributes/messages common to finish paths.
func ApplyLLMFinishAttributes(span trace.Span, invocation *LLMInvocation) {
	// Update span name
	span.SetName(GetLLMSpanName(invocation))

	// Collect all attributes
	var allAttrs []attribute.KeyValue
	allAttrs = append(allAttrs, GetLLMCommonAttributes(invocation)...)
	allAttrs = append(allAttrs, GetLLMRequestAttributes(invocation)...)
	allAttrs = append(allAttrs, GetLLMResponseAttributes(invocation)...)
	allAttrs = append(allAttrs, GetLLMMessageAttributesForSpan(invocation)...)

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			allAttrs = append(allAttrs, attribute.String(k, val))
		case int:
			allAttrs = append(allAttrs, attribute.Int(k, val))
		case int64:
			allAttrs = append(allAttrs, attribute.Int64(k, val))
		case float64:
			allAttrs = append(allAttrs, attribute.Float64(k, val))
		case bool:
			allAttrs = append(allAttrs, attribute.Bool(k, val))
		case []string:
			allAttrs = append(allAttrs, attribute.StringSlice(k, val))
		}
	}

	// Set all attributes on the span
	if len(allAttrs) > 0 {
		span.SetAttributes(allAttrs...)
	}
}

// ApplyErrorAttributes applies status and error attributes common to error paths.
func ApplyErrorAttributes(span trace.Span, err *Error) {
	span.SetStatus(codes.Error, err.Message)
	if span.IsRecording() {
		span.SetAttributes(attribute.String(AttrErrorType, err.Type))
	}
}

// ============================================================================
// Extended Span Utilities (LoongSuite Extension)
//
// The following functions provide span attribute handling for additional GenAI
// operations beyond basic LLM calls, including embeddings, tool execution,
// agent operations, document retrieval, and reranking.
// ============================================================================

// ApplyEmbeddingFinishAttributes applies attributes for embedding operations.
// (LoongSuite Extension)
func ApplyEmbeddingFinishAttributes(span trace.Span, invocation *EmbeddingInvocation) {
	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationEmbeddings),
		GenAISpanKind(SpanKindEmbedding),
	}

	if invocation.RequestModel != "" {
		attrs = append(attrs, GenAIRequestModel(invocation.RequestModel))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if invocation.InputCount != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIEmbeddingInputCount, *invocation.InputCount))
	}
	if invocation.Dimensions != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIEmbeddingDimensions, *invocation.Dimensions))
	}
	if invocation.InputTokens != nil {
		attrs = append(attrs, GenAIUsageInputTokens(*invocation.InputTokens))
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}

// ApplyExecuteToolFinishAttributes applies attributes for tool execution operations.
// (LoongSuite Extension)
func ApplyExecuteToolFinishAttributes(span trace.Span, invocation *ExecuteToolInvocation) {
	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationExecuteTool),
		GenAISpanKind(SpanKindTool),
		attribute.String(AttrGenAIToolName, invocation.ToolName),
	}

	if invocation.ToolCallID != "" {
		attrs = append(attrs, attribute.String(AttrGenAIToolCallID, invocation.ToolCallID))
	}

	// Add input/output as JSON if in experimental mode
	if IsExperimentalMode() && ShouldCaptureContentInSpan() {
		if invocation.Input != nil {
			attrs = append(attrs, attribute.String(AttrGenAIToolInput, MustJSONDumps(invocation.Input)))
		}
		if invocation.Output != nil {
			attrs = append(attrs, attribute.String(AttrGenAIToolOutput, MustJSONDumps(invocation.Output)))
		}
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}

// ApplyInvokeAgentFinishAttributes applies attributes for agent invocation operations.
// (LoongSuite Extension)
func ApplyInvokeAgentFinishAttributes(span trace.Span, invocation *InvokeAgentInvocation) {
	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationInvokeAgent),
		GenAISpanKind(SpanKindAgent),
	}

	if invocation.AgentName != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentName, invocation.AgentName))
	}
	if invocation.AgentID != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentID, invocation.AgentID))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}

	// Add messages if in experimental mode
	if IsExperimentalMode() && ShouldCaptureContentInSpan() {
		if len(invocation.InputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIInputMessages, InputMessagesToJSON(invocation.InputMessages)))
		}
		if len(invocation.OutputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIOutputMessages, OutputMessagesToJSON(invocation.OutputMessages)))
		}
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}

// ApplyCreateAgentFinishAttributes applies attributes for agent creation operations.
// (LoongSuite Extension)
func ApplyCreateAgentFinishAttributes(span trace.Span, invocation *CreateAgentInvocation) {
	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationCreateAgent),
		GenAISpanKind(SpanKindAgent),
	}

	if invocation.AgentName != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentName, invocation.AgentName))
	}
	if invocation.AgentID != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentID, invocation.AgentID))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}

// ApplyRetrieveFinishAttributes applies attributes for retrieve operations.
// (LoongSuite Extension)
func ApplyRetrieveFinishAttributes(span trace.Span, invocation *RetrieveInvocation) {
	attrs := []attribute.KeyValue{
		GenAISpanKind(SpanKindRetriever),
	}

	if invocation.Query != "" && IsExperimentalMode() && ShouldCaptureContentInSpan() {
		attrs = append(attrs, attribute.String(AttrGenAIRetrieveQuery, invocation.Query))
	}
	if invocation.TopK != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRetrieveTopK, *invocation.TopK))
	}
	if invocation.DocumentCount != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRetrieveDocumentCount, *invocation.DocumentCount))
	}
	if invocation.DataSourceName != "" {
		attrs = append(attrs, attribute.String(AttrGenAIRetrieveDataSourceName, invocation.DataSourceName))
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}

// ApplyRerankFinishAttributes applies attributes for rerank operations.
// (LoongSuite Extension)
func ApplyRerankFinishAttributes(span trace.Span, invocation *RerankInvocation) {
	attrs := []attribute.KeyValue{
		GenAISpanKind(SpanKindReranker),
	}

	if invocation.Query != "" && IsExperimentalMode() && ShouldCaptureContentInSpan() {
		attrs = append(attrs, attribute.String(AttrGenAIRerankQuery, invocation.Query))
	}
	if invocation.Model != "" {
		attrs = append(attrs, attribute.String(AttrGenAIRerankModel, invocation.Model))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if invocation.TopN != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRerankTopN, *invocation.TopN))
	}
	if invocation.InputCount != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRerankInputCount, *invocation.InputCount))
	}
	if invocation.OutputCount != nil {
		attrs = append(attrs, attribute.Int(AttrGenAIRerankOutputCount, *invocation.OutputCount))
	}

	// Add custom attributes
	for k, v := range invocation.Attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		}
	}

	span.SetAttributes(attrs...)
}
