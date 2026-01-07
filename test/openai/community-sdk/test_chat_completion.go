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

package main

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	// Create a mock HTTP server that simulates OpenAI API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"id": "chatcmpl-test123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I assist you today?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// Create OpenAI client pointing to mock server
	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = mockServer.URL
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	// Make a chat completion request (this will be instrumented)
	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello, how are you?",
			},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	})

	if err != nil {
		panic(err)
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "openai", "gpt-4")
		
		// Verify additional attributes
		span := stubs[0][0]
		temp := verifier.GetAttribute(span.Attributes, "gen_ai.request.temperature").AsFloat64()
		verifier.Assert(temp == 0.7, "Expected temperature to be 0.7, got %f", temp)
		
		maxTokens := verifier.GetAttribute(span.Attributes, "gen_ai.request.max_tokens").AsInt64()
		verifier.Assert(maxTokens == 100, "Expected max_tokens to be 100, got %d", maxTokens)
		
		// Verify usage tokens
		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 10, "Expected input tokens to be 10, got %d", inputTokens)
		
		outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
		verifier.Assert(outputTokens == 20, "Expected output tokens to be 20, got %d", outputTokens)
		
		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 30, "Expected total tokens to be 30, got %d", totalTokens)
		
		// Verify response ID
		responseID := verifier.GetAttribute(span.Attributes, "gen_ai.response.id").AsString()
		verifier.Assert(responseID == "chatcmpl-test123", "Expected response ID to be chatcmpl-test123, got %s", responseID)
		
		// Verify finish reason
		finishReasons := verifier.GetAttribute(span.Attributes, "gen_ai.response.finish_reasons").AsStringSlice()
		verifier.Assert(len(finishReasons) == 1 && finishReasons[0] == "stop", "Expected finish reason to be [stop], got %v", finishReasons)
	}, 1)
}
