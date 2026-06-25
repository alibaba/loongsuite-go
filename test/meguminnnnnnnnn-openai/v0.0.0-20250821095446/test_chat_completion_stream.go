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
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	openai "github.com/meguminnnnnnnnn/go-openai"
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
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}],"usage":null}` + "\n\n",
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" there!"},"finish_reason":null}],"usage":null}` + "\n\n",
			`data: {"id":"chatcmpl-stream123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":12,"total_tokens":20}}` + "\n\n",
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
	config.BaseURL = mockServer.URL + "/v1"
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
		Stream: true,
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
			fmt.Printf("Stream error: %v\n", err)
			break
		}
	}

	// Verify that the trace was captured correctly
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]

		// Verify operation name
		operationName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(operationName == "chat", "Expected operation name to be chat, got %s", operationName)

		// Verify system (127.0.0.1 maps to "local" in provider detection)
		system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
		verifier.Assert(system == "local", "Expected system to be local, got %s", system)

		// Verify model
		model := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(model == "gpt-4", "Expected model to be gpt-4, got %s", model)
	}, 1)
}
