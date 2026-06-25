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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/alibaba/loongsuite-go/test/verifier"
	deepseek "github.com/cohesion-org/deepseek-go"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// mockHTTPClient implements deepseek.HTTPDoer to intercept all HTTP requests
type mockHTTPClient struct {
	mockURL string
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Redirect the request to our mock server
	newReq, _ := http.NewRequestWithContext(req.Context(), req.Method, m.mockURL+req.URL.Path, req.Body)
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	return http.DefaultClient.Do(newReq)
}

func main() {
	// Create a mock HTTP server that simulates DeepSeek FIM API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and discard body
		io.Copy(io.Discard, r.Body)
		r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
			"id": "cmpl-fim-test123",
			"object": "text_completion",
			"created": 1677652288,
			"model": "deepseek-chat",
			"choices": [{
				"index": 0,
				"text": "    return a + b\n",
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 15,
				"completion_tokens": 8,
				"total_tokens": 23
			}
		}`
		_, _ = io.Copy(w, strings.NewReader(mockResponse))
	}))
	defer mockServer.Close()

	// Create DeepSeek client with custom HTTP client to redirect FIM requests
	client := deepseek.NewClient("test-api-key")
	client.HTTPClient = &mockHTTPClient{mockURL: mockServer.URL}

	ctx := context.Background()

	// Make a FIM completion request (this will be instrumented)
	_, err := client.CreateFIMCompletion(ctx, &deepseek.FIMCompletionRequest{
		Model:       deepseek.DeepSeekChat,
		Prompt:      "func add(a, b int) int {\n",
		Suffix:      "\n}\n",
		MaxTokens:   100,
		Temperature: 0.0,
	})

	if err != nil {
		panic(err)
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		// Verify span name: "text_completion deepseek-chat"
		verifier.Assert(span.Name == "text_completion deepseek-chat", "Expected span name to be 'text_completion deepseek-chat', got %s", span.Name)
		// Verify gen_ai.operation.name
		opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(opName == "text_completion", "Expected gen_ai.operation.name to be 'text_completion', got %s", opName)
		// Verify gen_ai.provider.name
		provider := verifier.GetAttribute(span.Attributes, "gen_ai.provider.name").AsString()
		verifier.Assert(provider == "deepseek", "Expected gen_ai.provider.name to be 'deepseek', got %s", provider)
		// Verify gen_ai.request.model
		model := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(model == "deepseek-chat", "Expected gen_ai.request.model to be 'deepseek-chat', got %s", model)
	}, 1)
}
