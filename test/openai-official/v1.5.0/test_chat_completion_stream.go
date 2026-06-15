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
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// Create a mock HTTP server that simulates OpenAI streaming API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Send streaming chunks
		chunks := []string{
			`data: {"id":"chatcmpl-stream-official123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"chatcmpl-stream-official123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" from official SDK!"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"chatcmpl-stream-official123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}` + "\n\n",
			"data: [DONE]\n\n",
		}

		for _, chunk := range chunks {
			fmt.Fprint(w, chunk)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer mockServer.Close()

	// Create OpenAI client pointing to mock server
	client := openai.NewClient(
		option.WithAPIKey("test-api-key"),
		option.WithBaseURL(mockServer.URL),
	)

	ctx := context.Background()

	// Make a streaming chat completion request (this will be instrumented)
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: shared.ChatModelGPT4,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello, how are you?"),
		},
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(100),
	})

	// Consume the stream
	for stream.Next() {
		_ = stream.Current()
	}

	if err := stream.Err(); err != nil {
		panic(err)
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "openai", "gpt-4")

		// Verify additional attributes
		span := stubs[0][0]
		temp := verifier.GetAttribute(span.Attributes, "gen_ai.request.temperature").AsFloat64()
		// Use approximate comparison for float values
		verifier.Assert(temp > 0.69 && temp < 0.71, "Expected temperature to be approximately 0.7, got %f", temp)

		maxTokens := verifier.GetAttribute(span.Attributes, "gen_ai.request.max_tokens").AsInt64()
		verifier.Assert(maxTokens == 100, "Expected max_tokens to be 100, got %d", maxTokens)
	}, 1)
}
