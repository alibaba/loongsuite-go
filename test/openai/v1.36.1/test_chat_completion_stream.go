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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" there!"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}` + "\n\n",
			"data: [DONE]\n\n",
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer mockServer.Close()

	// Create OpenAI client pointing to mock server
	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = mockServer.URL
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	// Make a streaming chat completion request (this will be instrumented)
	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Temperature: 0.8,
		Stream:      true,
	})

	if err != nil {
		panic(err)
	}
	defer stream.Close()

	// Consume the stream
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		xx, _ := json.Marshal(stubs)
		fmt.Println(string(xx))
		// For streaming, operation name should be "chat.stream"
		span := stubs[0][0]
		operationName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(operationName == "chat", "Expected operation name to be chat.stream, got %s", operationName)

		// Verify system
		system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
		verifier.Assert(system == "openai", "Expected system to be openai, got %s", system)

		// Verify model
		model := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(model == "gpt-4", "Expected model to be gpt-4, got %s", model)

		// Verify temperature
		temp := verifier.GetAttribute(span.Attributes, "gen_ai.request.temperature").AsFloat64()
		// Use approximate comparison for float values due to float32 -> float64 conversion
		verifier.Assert(temp > 0.79 && temp < 0.81, "Expected temperature to be approximately 0.8, got %f", temp)
	}, 1)
}
