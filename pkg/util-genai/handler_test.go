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
	"testing"
)

func TestNewTelemetryHandler(t *testing.T) {
	handler := NewTelemetryHandler()
	if handler == nil {
		t.Fatal("NewTelemetryHandler returned nil")
	}
	if handler.tracer == nil {
		t.Fatal("TelemetryHandler tracer is nil")
	}
}

func TestGetTelemetryHandler(t *testing.T) {
	handler := GetTelemetryHandler()
	if handler == nil {
		t.Fatal("GetTelemetryHandler returned nil")
	}

	// Should return the same instance
	handler2 := GetTelemetryHandler()
	if handler != handler2 {
		t.Fatal("GetTelemetryHandler should return singleton")
	}
}

func TestLLMInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	// Create an LLM invocation
	invocation := NewLLMInvocation("gpt-4")
	invocation.Provider = "openai"
	invocation.InputMessages = []InputMessage{
		{
			Role: "user",
			Parts: []MessagePart{
				Text{Content: "Hello, world!"},
			},
		},
	}

	// Start the invocation
	ctx = handler.StartLLM(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartLLM")
	}

	// Set output
	invocation.OutputMessages = []OutputMessage{
		{
			Role: "assistant",
			Parts: []MessagePart{
				Text{Content: "Hello! How can I help you?"},
			},
			FinishReason: FinishReasonStop,
		},
	}
	inputTokens := 10
	outputTokens := 8
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens

	// Stop the invocation
	handler.StopLLM(invocation)
}

func TestLLMInvocationWithError(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewLLMInvocation("gpt-4")
	invocation.Provider = "openai"

	ctx = handler.StartLLM(ctx, invocation)
	_ = ctx

	// Fail the invocation
	handler.FailLLM(invocation, &Error{
		Message: "API Error",
		Type:    "APIError",
	})
}

func TestEmbeddingInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewEmbeddingInvocation("text-embedding-3-small")
	invocation.Provider = "openai"
	inputCount := 5
	invocation.InputCount = &inputCount

	ctx = handler.StartEmbedding(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartEmbedding")
	}

	inputTokens := 100
	invocation.InputTokens = &inputTokens
	_ = ctx

	handler.StopEmbedding(invocation)
}

func TestExecuteToolInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewExecuteToolInvocation("get_weather")
	invocation.ToolCallID = "call_123"
	invocation.Input = map[string]any{"location": "San Francisco"}

	ctx = handler.StartExecuteTool(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartExecuteTool")
	}

	invocation.Output = map[string]any{"temperature": 72, "condition": "sunny"}
	_ = ctx

	handler.StopExecuteTool(invocation)
}

func TestInvokeAgentInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewInvokeAgentInvocation()
	invocation.AgentName = "assistant"
	invocation.Provider = "openai"

	ctx = handler.StartInvokeAgent(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartInvokeAgent")
	}
	_ = ctx

	handler.StopInvokeAgent(invocation)
}

func TestCreateAgentInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewCreateAgentInvocation()
	invocation.AgentName = "my-agent"
	invocation.Provider = "openai"

	ctx = handler.StartCreateAgent(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartCreateAgent")
	}

	invocation.AgentID = "agent_123"
	_ = ctx

	handler.StopCreateAgent(invocation)
}

func TestRetrieveInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewRetrieveInvocation()
	invocation.Query = "What is OpenTelemetry?"
	topK := 10
	invocation.TopK = &topK

	ctx = handler.StartRetrieve(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartRetrieve")
	}

	docCount := 5
	invocation.DocumentCount = &docCount
	_ = ctx

	handler.StopRetrieve(invocation)
}

func TestRerankInvocation(t *testing.T) {
	handler := NewTelemetryHandler()
	ctx := context.Background()

	invocation := NewRerankInvocation()
	invocation.Query = "What is OpenTelemetry?"
	invocation.Model = "rerank-english-v2.0"
	invocation.Provider = "cohere"
	inputCount := 10
	invocation.InputCount = &inputCount

	ctx = handler.StartRerank(ctx, invocation)
	if invocation.span == nil {
		t.Fatal("Span should be set after StartRerank")
	}

	outputCount := 5
	invocation.OutputCount = &outputCount
	_ = ctx

	handler.StopRerank(invocation)
}
