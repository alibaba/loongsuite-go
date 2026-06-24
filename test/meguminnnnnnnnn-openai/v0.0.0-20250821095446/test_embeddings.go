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
	openai "github.com/meguminnnnnnnnn/go-openai"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	// Create a mock HTTP server that simulates OpenAI embeddings API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"object": "list",
			"data": [{
				"object": "embedding",
				"embedding": [0.0023064255, -0.009327292, 0.015797347],
				"index": 0
			}],
			"model": "text-embedding-ada-002",
			"usage": {
				"prompt_tokens": 5,
				"completion_tokens": 0,
				"total_tokens": 5
			}
		}`
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// Create OpenAI client pointing to mock server
	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = mockServer.URL + "/v1"
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	// Make an embeddings request (this will be instrumented)
	_, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: []string{"Hello world"},
		Model: openai.AdaEmbeddingV2,
	})

	if err != nil {
		panic(err)
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]

		// Verify operation name
		operationName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(operationName == "embeddings", "Expected operation name to be embeddings, got %s", operationName)

		// Verify system (127.0.0.1 maps to "local" in provider detection)
		system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
		verifier.Assert(system == "local", "Expected system to be local, got %s", system)

		// Verify model
		model := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(model == "text-embedding-ada-002", "Expected model to be text-embedding-ada-002, got %s", model)
	}, 1)
}
