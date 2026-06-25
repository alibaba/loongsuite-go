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

	"github.com/alibaba/loongsuite-go/test/verifier"
	deepseek "github.com/cohesion-org/deepseek-go"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	// Create a mock HTTP server that simulates DeepSeek API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"id": "chatcmpl-deepseek-test123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "deepseek-chat",
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

	// Create DeepSeek client pointing to mock server
	client := deepseek.NewClient("test-api-key", mockServer.URL+"/")

	ctx := context.Background()

	// Make a chat completion request (this will be instrumented)
	_, err := client.CreateChatCompletion(ctx, &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{
				Role:    deepseek.ChatMessageRoleUser,
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
		span := stubs[0][0]
		// Verify span name: "chat deepseek-chat"
		verifier.Assert(span.Name == "chat deepseek-chat", "Expected span name to be 'chat deepseek-chat', got %s", span.Name)
		// Verify gen_ai.operation.name
		opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(opName == "chat", "Expected gen_ai.operation.name to be 'chat', got %s", opName)
		// Verify gen_ai.provider.name
		provider := verifier.GetAttribute(span.Attributes, "gen_ai.provider.name").AsString()
		verifier.Assert(provider == "deepseek", "Expected gen_ai.provider.name to be 'deepseek', got %s", provider)
		// Verify gen_ai.request.model
		model := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(model == "deepseek-chat", "Expected gen_ai.request.model to be 'deepseek-chat', got %s", model)
	}, 1)
}
