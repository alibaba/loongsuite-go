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
	// Create a mock HTTP server that simulates the Gemini API EmbedContent endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"embeddings": [
				{
					"values": [0.1, 0.2, 0.3, 0.4, 0.5]
				}
			]
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

	// Call EmbedContent
	model := "text-embedding-004"
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Hello, world!"},
			},
			Role: "user",
		},
	}

	_, err = client.Models.EmbedContent(ctx, model, contents, nil)
	if err != nil {
		panic(fmt.Sprintf("EmbedContent failed: %v", err))
	}

	// Verify traces
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "embeddings", "google_genai", "text-embedding-004")

		span := stubs[0][0]

		// Verify gen_ai.span.kind is EMBEDDING
		spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
		verifier.Assert(spanKind == "EMBEDDING", "Expected gen_ai.span.kind to be EMBEDDING, got %s", spanKind)

		// Verify provider name
		providerName := verifier.GetAttribute(span.Attributes, "gen_ai.provider.name").AsString()
		verifier.Assert(providerName == "google", "Expected gen_ai.provider.name to be google, got %s", providerName)
	}, 1)
}
