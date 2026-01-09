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
	"github.com/openai/openai-go/shared"
	"net/http"
	"net/http/httptest"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	// Create a mock HTTP server that simulates OpenAI API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
"id": "chatcmpl-official-test123",
"object": "chat.completion",
"created": 1677652288,
"model": "gpt-4",
"choices": [{
"index": 0,
"message": {
"role": "assistant",
"content": "Hello from official SDK!"
},
"finish_reason": "stop"
}],
"usage": {
"prompt_tokens": 15,
"completion_tokens": 25,
"total_tokens": 40
}
}`
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// Create OpenAI client pointing to mock server
	client := openai.NewClient(
		option.WithAPIKey("test-api-key"),
		option.WithBaseURL(mockServer.URL),
	)

	ctx := context.Background()

	// Make a chat completion request (this will be instrumented)
	_, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: shared.ChatModelGPT4,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello, how are you?"),
		},
		Temperature: openai.Float(0.8),
		MaxTokens:   openai.Int(150),
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
		// Use approximate comparison for float values
		verifier.Assert(temp > 0.79 && temp < 0.81, "Expected temperature to be approximately 0.8, got %f", temp)

		maxTokens := verifier.GetAttribute(span.Attributes, "gen_ai.request.max_tokens").AsInt64()
		verifier.Assert(maxTokens == 150, "Expected max_tokens to be 150, got %d", maxTokens)

		// Verify usage tokens
		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 15, "Expected input tokens to be 15, got %d", inputTokens)

		outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
		verifier.Assert(outputTokens == 25, "Expected output tokens to be 25, got %d", outputTokens)

		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 40, "Expected total tokens to be 40, got %d", totalTokens)

		// Verify response ID
		responseID := verifier.GetAttribute(span.Attributes, "gen_ai.response.id").AsString()
		verifier.Assert(responseID == "chatcmpl-official-test123", "Expected response ID to be chatcmpl-official-test123, got %s", responseID)

		// Verify finish reason
		finishReasons := verifier.GetAttribute(span.Attributes, "gen_ai.response.finish_reasons").AsStringSlice()
		verifier.Assert(len(finishReasons) == 1 && finishReasons[0] == "stop", "Expected finish reason to be [stop], got %v", finishReasons)
	}, 1)
}
