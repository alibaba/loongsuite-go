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
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/genai"
)

func main() {
	// Create a mock HTTP server that simulates the Gemini API GenerateContent endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"candidates": [
				{
					"content": {
						"parts": [{"text": "Hello! I'm a mock Gemini response."}],
						"role": "model"
					},
					"finishReason": "STOP",
					"index": 0
				}
			],
			"usageMetadata": {
				"promptTokenCount": 15,
				"candidatesTokenCount": 25,
				"totalTokenCount": 40
			},
			"modelVersion": "gemini-2.0-flash",
			"responseId": "test-response-id-123"
		}`
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	ctx := context.Background()

	// Create GenAI client with mock server
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  "test-api-key",
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			BaseURL: mockServer.URL,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Call GenerateContent
	model := "gemini-2.0-flash"
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Hello, how are you?"},
			},
			Role: "user",
		},
	}

	_, err = client.Models.GenerateContent(ctx, model, contents, nil)
	if err != nil {
		panic(fmt.Sprintf("GenerateContent failed: %v", err))
	}

	// Verify traces
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "google_genai", "gemini-2.0-flash")

		span := stubs[0][0]

		// Verify token usage attributes
		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 15, "Expected input tokens to be 15, got %d", inputTokens)

		outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
		verifier.Assert(outputTokens == 25, "Expected output tokens to be 25, got %d", outputTokens)

		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 40, "Expected total tokens to be 40, got %d", totalTokens)

		// Verify response ID
		responseID := verifier.GetAttribute(span.Attributes, "gen_ai.response.id").AsString()
		verifier.Assert(responseID == "test-response-id-123", "Expected response ID to be test-response-id-123, got %s", responseID)

		// Verify finish reason
		finishReasons := verifier.GetAttribute(span.Attributes, "gen_ai.response.finish_reasons").AsStringSlice()
		verifier.Assert(len(finishReasons) == 1 && finishReasons[0] == "STOP", "Expected finish reason to be [STOP], got %v", finishReasons)

		// Verify response model
		responseModel := verifier.GetAttribute(span.Attributes, "gen_ai.response.model").AsString()
		verifier.Assert(responseModel == "gemini-2.0-flash", "Expected response model to be gemini-2.0-flash, got %s", responseModel)
	}, 1)
}
