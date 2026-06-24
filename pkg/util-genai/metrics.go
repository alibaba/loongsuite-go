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
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Default bucket boundaries for GenAI metrics
var (
	// Duration buckets in seconds
	GenAIClientOperationDurationBuckets = []float64{
		0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.64, 1.28,
		2.56, 5.12, 10.24, 20.48, 40.96, 81.92,
	}

	// Token usage buckets
	GenAIClientTokenUsageBuckets = []float64{
		1, 4, 16, 64, 256, 1024, 4096, 16384,
		65536, 262144, 1048576, 4194304, 16777216, 67108864,
	}
)

// MetricsRecorder records duration and token usage metrics for GenAI invocations.
type MetricsRecorder struct {
	durationHistogram metric.Float64Histogram
	tokenHistogram    metric.Int64Histogram
}

// NewMetricsRecorder creates a new MetricsRecorder with the given meter.
func NewMetricsRecorder(meter metric.Meter) (*MetricsRecorder, error) {
	durationHist, err := meter.Float64Histogram(
		MetricGenAIClientOperationDuration,
		metric.WithDescription("Duration of GenAI client operation"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(GenAIClientOperationDurationBuckets...),
	)
	if err != nil {
		return nil, err
	}

	tokenHist, err := meter.Int64Histogram(
		MetricGenAIClientTokenUsage,
		metric.WithDescription("Number of input and output tokens used by GenAI clients"),
		metric.WithUnit("{token}"),
		metric.WithExplicitBucketBoundaries(GenAIClientTokenUsageBuckets...),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsRecorder{
		durationHistogram: durationHist,
		tokenHistogram:    tokenHist,
	}, nil
}

// ============================================================================
// Basic LLM Metrics
//
// These methods record duration and token usage metrics for LLM operations
// (chat, text_completion, generate_content), which is the core functionality
// for instrumenting language model calls.
// ============================================================================

// RecordLLM records duration and token metrics for an LLM invocation.
func (r *MetricsRecorder) RecordLLM(ctx context.Context, invocation *LLMInvocation, duration time.Duration, errorType string) {
	if r == nil {
		return
	}

	// Build common attributes
	attrs := []attribute.KeyValue{
		GenAIOperationName(invocation.OperationName),
	}

	if invocation.RequestModel != "" {
		attrs = append(attrs, GenAIRequestModel(invocation.RequestModel))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if invocation.ResponseModelName != "" {
		attrs = append(attrs, GenAIResponseModel(invocation.ResponseModelName))
	}
	if errorType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errorType))
	}

	// Record duration
	if duration > 0 {
		r.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Record token usage
	if invocation.InputTokens != nil {
		tokenAttrs := make([]attribute.KeyValue, len(attrs)+1)
		copy(tokenAttrs, attrs)
		tokenAttrs[len(attrs)] = attribute.String(AttrGenAITokenType, string(TokenTypeInput))
		r.tokenHistogram.Record(ctx, int64(*invocation.InputTokens), metric.WithAttributes(tokenAttrs...))
	}
	if invocation.OutputTokens != nil {
		tokenAttrs := make([]attribute.KeyValue, len(attrs)+1)
		copy(tokenAttrs, attrs)
		tokenAttrs[len(attrs)] = attribute.String(AttrGenAITokenType, string(TokenTypeOutput))
		r.tokenHistogram.Record(ctx, int64(*invocation.OutputTokens), metric.WithAttributes(tokenAttrs...))
	}
}

// ============================================================================
// Extended Metrics (LoongSuite Extension)
//
// The following methods record metrics for additional GenAI operations beyond
// basic LLM calls, including embeddings, tool execution, and agent operations.
// ============================================================================

// RecordEmbedding records metrics for an embedding invocation.
// (LoongSuite Extension)
func (r *MetricsRecorder) RecordEmbedding(ctx context.Context, invocation *EmbeddingInvocation, duration time.Duration, errorType string) {
	if r == nil {
		return
	}

	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationEmbeddings),
	}

	if invocation.RequestModel != "" {
		attrs = append(attrs, GenAIRequestModel(invocation.RequestModel))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if errorType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errorType))
	}

	// Record duration
	if duration > 0 {
		r.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}

	// Record token usage
	if invocation.InputTokens != nil {
		tokenAttrs := make([]attribute.KeyValue, len(attrs)+1)
		copy(tokenAttrs, attrs)
		tokenAttrs[len(attrs)] = attribute.String(AttrGenAITokenType, string(TokenTypeInput))
		r.tokenHistogram.Record(ctx, int64(*invocation.InputTokens), metric.WithAttributes(tokenAttrs...))
	}
}

// RecordTool records metrics for a tool execution invocation.
// (LoongSuite Extension)
func (r *MetricsRecorder) RecordTool(ctx context.Context, invocation *ExecuteToolInvocation, duration time.Duration, errorType string) {
	if r == nil {
		return
	}

	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationExecuteTool),
		attribute.String(AttrGenAIToolName, invocation.ToolName),
	}

	if errorType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errorType))
	}

	// Record duration
	if duration > 0 {
		r.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}
}

// RecordAgent records metrics for an agent invocation.
// (LoongSuite Extension)
func (r *MetricsRecorder) RecordAgent(ctx context.Context, invocation *InvokeAgentInvocation, duration time.Duration, errorType string) {
	if r == nil {
		return
	}

	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationInvokeAgent),
	}

	if invocation.AgentName != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentName, invocation.AgentName))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if errorType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errorType))
	}

	// Record duration
	if duration > 0 {
		r.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	}
}
