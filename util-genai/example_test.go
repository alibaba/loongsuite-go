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

package utilgenai_test

import (
	"context"
	"fmt"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"
)

// ExampleTelemetryHandler_StartLLM demonstrates a basic LLM invocation lifecycle
// using the TelemetryHandler. This is the most common usage pattern for instrumenting
// LLM calls such as chat completions.
func ExampleTelemetryHandler_StartLLM() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// Create an LLM invocation with the request model
	invocation := utilgenai.NewLLMInvocation("gpt-4")
	invocation.Provider = "openai"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "What is OpenTelemetry?"},
			},
		},
	}

	// Start the invocation - this creates a new span
	ctx = handler.StartLLM(ctx, invocation)

	// Simulate receiving a response from the LLM
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "OpenTelemetry is an observability framework..."},
			},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inputTokens := 12
	outputTokens := 45
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens
	invocation.ResponseModelName = "gpt-4-0613"
	invocation.ResponseID = "chatcmpl-abc123"

	// Stop the invocation - this ends the span and records metrics
	handler.StopLLM(invocation)

	_ = ctx
	fmt.Println("LLM invocation completed successfully")
	// Output: LLM invocation completed successfully
}

// ExampleTelemetryHandler_FailLLM demonstrates error handling for LLM invocations.
// When an LLM call fails, use FailLLM to record the error in the span.
func ExampleTelemetryHandler_FailLLM() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	invocation := utilgenai.NewLLMInvocation("gpt-4")
	invocation.Provider = "openai"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Hello!"},
			},
		},
	}

	// Start the invocation
	ctx = handler.StartLLM(ctx, invocation)

	// Simulate an error from the LLM API
	apiErr := &utilgenai.Error{
		Message: "Rate limit exceeded",
		Type:    "RateLimitError",
	}

	// Fail the invocation - this records the error and ends the span
	handler.FailLLM(invocation, apiErr)

	_ = ctx
	fmt.Println("LLM invocation failed with error recorded")
	// Output: LLM invocation failed with error recorded
}

// ExampleTelemetryHandler_StartEmbedding demonstrates instrumenting an embedding operation.
// Use this for text embedding API calls.
func ExampleTelemetryHandler_StartEmbedding() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// Create an embedding invocation
	invocation := utilgenai.NewEmbeddingInvocation("text-embedding-3-small")
	invocation.Provider = "openai"
	inputCount := 3
	invocation.InputCount = &inputCount

	// Start the embedding invocation
	ctx = handler.StartEmbedding(ctx, invocation)

	// Simulate receiving embedding results
	inputTokens := 150
	invocation.InputTokens = &inputTokens
	dimensions := 1536
	invocation.DimensionCount = &dimensions

	// Stop the embedding invocation
	handler.StopEmbedding(invocation)

	_ = ctx
	fmt.Println("Embedding invocation completed successfully")
	// Output: Embedding invocation completed successfully
}

// ExampleTelemetryHandler_StartExecuteTool demonstrates instrumenting a tool execution.
// Use this when an LLM requests a tool/function call and you execute it.
func ExampleTelemetryHandler_StartExecuteTool() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// Create a tool execution invocation
	invocation := utilgenai.NewExecuteToolInvocation("get_weather")
	invocation.ToolCallID = "call_abc123"
	invocation.Input = map[string]any{
		"location": "San Francisco",
		"unit":     "celsius",
	}

	// Start the tool execution
	ctx = handler.StartExecuteTool(ctx, invocation)

	// Simulate executing the tool and getting results
	invocation.Output = map[string]any{
		"temperature": 18,
		"condition":   "partly cloudy",
		"humidity":    65,
	}

	// Stop the tool execution
	handler.StopExecuteTool(invocation)

	_ = ctx
	fmt.Println("Tool execution completed successfully")
	// Output: Tool execution completed successfully
}

// ExampleTelemetryHandler_StartInvokeAgent demonstrates instrumenting an agent invocation.
// Use this when invoking an AI agent that may perform multiple steps.
func ExampleTelemetryHandler_StartInvokeAgent() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// Create an agent invocation
	invocation := utilgenai.NewInvokeAgentInvocation()
	invocation.AgentName = "research-assistant"
	invocation.Provider = "openai"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Research the latest trends in AI observability"},
			},
		},
	}

	// Start the agent invocation
	ctx = handler.StartInvokeAgent(ctx, invocation)

	// Simulate agent completing its work
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Here are the latest trends in AI observability..."},
			},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}

	// Stop the agent invocation
	handler.StopInvokeAgent(invocation)

	_ = ctx
	fmt.Println("Agent invocation completed successfully")
	// Output: Agent invocation completed successfully
}

// ExampleNewLLMInvocation demonstrates creating an LLM invocation with
// various request parameters configured.
func ExampleNewLLMInvocation() {
	invocation := utilgenai.NewLLMInvocation("gpt-4-turbo")
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat

	// Set optional request parameters
	temperature := 0.7
	invocation.Temperature = &temperature
	topP := 0.9
	invocation.TopP = &topP
	maxTokens := 2048
	invocation.MaxTokens = &maxTokens
	invocation.StopSequences = []string{"END", "STOP"}

	// Set system instruction
	invocation.SystemInstruction = []utilgenai.MessagePart{
		utilgenai.Text{Content: "You are a helpful assistant."},
	}

	// Set tool definitions
	invocation.ToolDefinitions = []utilgenai.FunctionToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City name",
					},
				},
			},
		},
	}

	// Set input messages
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "What's the weather in Tokyo?"},
			},
		},
	}

	fmt.Printf("Model: %s, Provider: %s, Operation: %s\n",
		invocation.RequestModel, invocation.Provider, invocation.OperationName)
	// Output: Model: gpt-4-turbo, Provider: openai, Operation: chat
}

// ExampleTelemetryHandler_deferPattern demonstrates the recommended defer pattern
// for handling both success and error cases in LLM invocations.
func ExampleTelemetryHandler_deferPattern() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	invocation := utilgenai.NewLLMInvocation("gpt-4")
	invocation.Provider = "openai"

	ctx = handler.StartLLM(ctx, invocation)

	// Use a closure to simulate the typical defer pattern
	var callErr error
	defer func() {
		if callErr != nil {
			handler.FailLLM(invocation, &utilgenai.Error{
				Message: callErr.Error(),
				Type:    "APIError",
			})
		} else {
			handler.StopLLM(invocation)
		}
	}()

	// Simulate a successful call
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Hello!"},
			},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inputTokens := 5
	outputTokens := 3
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens

	_ = ctx
	fmt.Println("Defer pattern example completed")
	// Output: Defer pattern example completed
}

// ExampleTelemetryHandler_streaming demonstrates instrumenting a streaming LLM call
// with gen_ai.request.stream and gen_ai.response.time_to_first_chunk.
func ExampleTelemetryHandler_streaming() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// Create a streaming LLM invocation
	invocation := utilgenai.NewLLMInvocation("gpt-4o")
	invocation.Provider = "openai"
	stream := true
	invocation.Stream = &stream
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Explain quantum computing briefly."},
			},
		},
	}

	// Start the invocation
	ctx = handler.StartLLM(ctx, invocation)

	// Simulate streaming - record time to first chunk
	timeToFirstChunk := 0.35 // 350ms to first chunk
	invocation.TimeToFirstChunk = &timeToFirstChunk

	// After collecting all chunks, set the complete response
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Quantum computing uses quantum bits..."},
			},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inputTokens := 8
	outputTokens := 42
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens

	handler.StopLLM(invocation)

	_ = ctx
	fmt.Println("Streaming invocation completed")
	// Output: Streaming invocation completed
}

// ExampleTelemetryHandler_conversationTracking demonstrates using
// gen_ai.conversation.id to track multi-turn conversations.
func ExampleTelemetryHandler_conversationTracking() {
	handler := utilgenai.NewTelemetryHandler()
	ctx := context.Background()

	// First turn in a conversation
	invocation := utilgenai.NewLLMInvocation("gpt-4o")
	invocation.Provider = "openai"
	invocation.ConversationID = "conv_5j66UpCpwteGg4YSxUnt7lPY"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "What is the capital of France?"},
			},
		},
	}

	ctx = handler.StartLLM(ctx, invocation)

	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "The capital of France is Paris."},
			},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inputTokens := 10
	outputTokens := 8
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens

	handler.StopLLM(invocation)

	_ = ctx
	fmt.Println("Conversation turn completed")
	// Output: Conversation turn completed
}
