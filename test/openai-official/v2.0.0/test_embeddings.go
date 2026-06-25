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
	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// Create a mock HTTP server that simulates OpenAI Embeddings API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
"object": "list",
"data": [
    {
        "object": "embedding",
        "embedding": [0.1, 0.2, 0.3],
        "index": 0
    }
],
"model": "text-embedding-3-small",
"usage": {
    "prompt_tokens": 5,
    "total_tokens": 5
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

	// Make an embedding request (this will be instrumented)
	_, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String("Hello world"),
		},
		Model: openai.EmbeddingModelTextEmbedding3Small,
	})

	if err != nil {
		panic(err)
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]

		// Verify span name and core attributes
		verifier.Assert(span.Name == "embeddings text-embedding-3-small", "Expected span name 'embeddings text-embedding-3-small', got %s", span.Name)
		verifier.Assert(span.SpanKind == trace.SpanKindClient, "Expected client span, got %d", span.SpanKind)

		// Verify gen_ai attributes
		providerName := verifier.GetAttribute(span.Attributes, "gen_ai.provider.name").AsString()
		verifier.Assert(providerName == "openai", "Expected provider name 'openai', got %s", providerName)

		opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(opName == "embeddings", "Expected operation name 'embeddings', got %s", opName)

		reqModel := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(reqModel == "text-embedding-3-small", "Expected request model 'text-embedding-3-small', got %s", reqModel)

		// Verify usage tokens
		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 5, "Expected input tokens to be 5, got %d", inputTokens)

		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 5, "Expected total tokens to be 5, got %d", totalTokens)

		// Verify response model
		respModel := verifier.GetAttribute(span.Attributes, "gen_ai.response.model").AsString()
		verifier.Assert(respModel == "text-embedding-3-small", "Expected response model 'text-embedding-3-small', got %s", respModel)
	}, 1)
}
