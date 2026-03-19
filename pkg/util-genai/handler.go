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
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName    = "github.com/alibaba/loongsuite-go-agent/pkg/util-genai"
	instrumentationVersion = "0.1.0"
)

// TelemetryHandler manages GenAI invocation lifecycles and emits telemetry data.
// It supports starting, stopping, and failing various GenAI operations.
//
// Basic operations:
//   - LLM/Chat: StartLLM, StopLLM, FailLLM
//
// Extended operations (LoongSuite Extension):
//   - Embedding: StartEmbedding, StopEmbedding, FailEmbedding
//   - Tool Execution: StartExecuteTool, StopExecuteTool, FailExecuteTool
//   - Agent Invocation: StartInvokeAgent, StopInvokeAgent, FailInvokeAgent
//   - Agent Creation: StartCreateAgent, StopCreateAgent, FailCreateAgent
//   - Document Retrieval: StartRetrieve, StopRetrieve, FailRetrieve
//   - Document Reranking: StartRerank, StopRerank, FailRerank
type TelemetryHandler struct {
	tracer          trace.Tracer
	metricsRecorder *MetricsRecorder
}

// TelemetryHandlerOption is a function that configures a TelemetryHandler.
type TelemetryHandlerOption func(*telemetryHandlerConfig)

type telemetryHandlerConfig struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
}

// WithTracerProvider sets the tracer provider for the TelemetryHandler.
func WithTracerProvider(tp trace.TracerProvider) TelemetryHandlerOption {
	return func(c *telemetryHandlerConfig) {
		c.tracerProvider = tp
	}
}

// WithMeterProvider sets the meter provider for the TelemetryHandler.
func WithMeterProvider(mp metric.MeterProvider) TelemetryHandlerOption {
	return func(c *telemetryHandlerConfig) {
		c.meterProvider = mp
	}
}

// NewTelemetryHandler creates a new TelemetryHandler with the given options.
func NewTelemetryHandler(opts ...TelemetryHandlerOption) *TelemetryHandler {
	cfg := &telemetryHandlerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var tracer trace.Tracer
	if cfg.tracerProvider != nil {
		tracer = cfg.tracerProvider.Tracer(instrumentationName, trace.WithInstrumentationVersion(instrumentationVersion))
	} else {
		tracer = otel.Tracer(instrumentationName, trace.WithInstrumentationVersion(instrumentationVersion))
	}

	var metricsRecorder *MetricsRecorder
	var meter metric.Meter
	if cfg.meterProvider != nil {
		meter = cfg.meterProvider.Meter(instrumentationName, metric.WithInstrumentationVersion(instrumentationVersion))
	} else {
		meter = otel.Meter(instrumentationName, metric.WithInstrumentationVersion(instrumentationVersion))
	}
	metricsRecorder, _ = NewMetricsRecorder(meter)

	return &TelemetryHandler{
		tracer:          tracer,
		metricsRecorder: metricsRecorder,
	}
}

// Singleton handler instance
var (
	defaultHandler     *TelemetryHandler
	defaultHandlerOnce sync.Once
)

// GetTelemetryHandler returns a singleton TelemetryHandler instance.
func GetTelemetryHandler(opts ...TelemetryHandlerOption) *TelemetryHandler {
	defaultHandlerOnce.Do(func() {
		defaultHandler = NewTelemetryHandler(opts...)
	})
	return defaultHandler
}

// ============================================================================
// Basic LLM Operations
//
// These methods handle the lifecycle of LLM invocations (chat, text_completion,
// generate_content), which is the core functionality for instrumenting language model calls.
// ============================================================================

// StartLLM starts an LLM invocation and creates a pending span entry.
func (h *TelemetryHandler) StartLLM(ctx context.Context, invocation *LLMInvocation) context.Context {
	spanName := GetLLMSpanName(invocation)
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopLLM finalizes an LLM invocation successfully and ends its span.
func (h *TelemetryHandler) StopLLM(invocation *LLMInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyLLMFinishAttributes(invocation.span, invocation)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordLLM(invocation.ctx, invocation, duration, "")
	}

	invocation.span.End()
}

// FailLLM fails an LLM invocation and ends its span with error status.
func (h *TelemetryHandler) FailLLM(invocation *LLMInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyLLMFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordLLM(invocation.ctx, invocation, duration, err.Type)
	}

	invocation.span.End()
}

// ============================================================================
// Extended Operations (LoongSuite Extension)
//
// The following methods provide instrumentation for additional GenAI operations
// beyond basic LLM calls, including embeddings, tool execution, agent operations,
// document retrieval, and reranking.
// ============================================================================

// StartEmbedding starts an embedding invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartEmbedding(ctx context.Context, invocation *EmbeddingInvocation) context.Context {
	spanName := fmt.Sprintf("%s %s", OperationEmbeddings, invocation.RequestModel)
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopEmbedding finalizes an embedding invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopEmbedding(invocation *EmbeddingInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyEmbeddingFinishAttributes(invocation.span, invocation)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordEmbedding(invocation.ctx, invocation, duration, "")
	}

	invocation.span.End()
}

// FailEmbedding fails an embedding invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailEmbedding(invocation *EmbeddingInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyEmbeddingFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordEmbedding(invocation.ctx, invocation, duration, err.Type)
	}

	invocation.span.End()
}

// StartExecuteTool starts a tool execution invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartExecuteTool(ctx context.Context, invocation *ExecuteToolInvocation) context.Context {
	spanName := fmt.Sprintf("%s %s", OperationExecuteTool, invocation.ToolName)
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopExecuteTool finalizes a tool execution invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopExecuteTool(invocation *ExecuteToolInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyExecuteToolFinishAttributes(invocation.span, invocation)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordTool(invocation.ctx, invocation, duration, "")
	}

	invocation.span.End()
}

// FailExecuteTool fails a tool execution invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailExecuteTool(invocation *ExecuteToolInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyExecuteToolFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordTool(invocation.ctx, invocation, duration, err.Type)
	}

	invocation.span.End()
}

// StartInvokeAgent starts an agent invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartInvokeAgent(ctx context.Context, invocation *InvokeAgentInvocation) context.Context {
	var spanName string
	if invocation.AgentName != "" {
		spanName = fmt.Sprintf("%s %s", OperationInvokeAgent, invocation.AgentName)
	} else {
		spanName = string(OperationInvokeAgent)
	}
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopInvokeAgent finalizes an agent invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopInvokeAgent(invocation *InvokeAgentInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyInvokeAgentFinishAttributes(invocation.span, invocation)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordAgent(invocation.ctx, invocation, duration, "")
	}

	invocation.span.End()
}

// FailInvokeAgent fails an agent invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailInvokeAgent(invocation *InvokeAgentInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyInvokeAgentFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)

	// Calculate duration
	endTime := float64(time.Now().UnixNano()) / 1e9
	duration := time.Duration((endTime - invocation.monotonicStartS) * float64(time.Second))

	// Record metrics
	if h.metricsRecorder != nil && invocation.ctx != nil {
		h.metricsRecorder.RecordAgent(invocation.ctx, invocation, duration, err.Type)
	}

	invocation.span.End()
}

// StartCreateAgent starts an agent creation invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartCreateAgent(ctx context.Context, invocation *CreateAgentInvocation) context.Context {
	var spanName string
	if invocation.AgentName != "" {
		spanName = fmt.Sprintf("%s %s", OperationCreateAgent, invocation.AgentName)
	} else {
		spanName = string(OperationCreateAgent)
	}
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopCreateAgent finalizes an agent creation invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopCreateAgent(invocation *CreateAgentInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyCreateAgentFinishAttributes(invocation.span, invocation)
	invocation.span.End()
}

// FailCreateAgent fails an agent creation invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailCreateAgent(invocation *CreateAgentInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyCreateAgentFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)
	invocation.span.End()
}

// StartRetrieve starts a retrieve documents invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartRetrieve(ctx context.Context, invocation *RetrieveInvocation) context.Context {
	spanName := "retrieve_documents"
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopRetrieve finalizes a retrieve documents invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopRetrieve(invocation *RetrieveInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyRetrieveFinishAttributes(invocation.span, invocation)
	invocation.span.End()
}

// FailRetrieve fails a retrieve documents invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailRetrieve(invocation *RetrieveInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyRetrieveFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)
	invocation.span.End()
}

// StartRerank starts a rerank documents invocation and creates a pending span entry.
// (LoongSuite Extension)
func (h *TelemetryHandler) StartRerank(ctx context.Context, invocation *RerankInvocation) context.Context {
	spanName := "rerank_documents"
	newCtx, span := h.tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))

	invocation.span = span
	invocation.ctx = newCtx
	invocation.monotonicStartS = float64(time.Now().UnixNano()) / 1e9

	return newCtx
}

// StopRerank finalizes a rerank documents invocation successfully and ends its span.
// (LoongSuite Extension)
func (h *TelemetryHandler) StopRerank(invocation *RerankInvocation) {
	if invocation.span == nil {
		return
	}

	ApplyRerankFinishAttributes(invocation.span, invocation)
	invocation.span.End()
}

// FailRerank fails a rerank documents invocation and ends its span with error status.
// (LoongSuite Extension)
func (h *TelemetryHandler) FailRerank(invocation *RerankInvocation, err *Error) {
	if invocation.span == nil {
		return
	}

	ApplyRerankFinishAttributes(invocation.span, invocation)
	ApplyErrorAttributes(invocation.span, err)
	invocation.span.End()
}
