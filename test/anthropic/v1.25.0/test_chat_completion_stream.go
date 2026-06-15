// Copyright (c) 2026 Alibaba Group Holding Ltd.
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
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		chunks := []string{
			`data: {"type":"message_start","message":{"id":"msg_stream_123","type":"message","role":"assistant","model":"claude-3-7-sonnet-latest","content":[],"stop_reason":"","stop_sequence":"","container":{"id":"container_stream","expires_at":"2026-01-01T00:00:00Z"},"usage":{"cache_creation":{"ephemeral_1h_input_tokens":0,"ephemeral_5m_input_tokens":0},"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"inference_geo":"us","input_tokens":10,"output_tokens":0,"server_tool_use":{"web_fetch_requests":0,"web_search_requests":0},"service_tier":"standard"}}}` + "\n\n",
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}` + "\n\n",
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" from anthropic stream"}}` + "\n\n",
			`data: {"type":"content_block_stop","index":0}` + "\n\n",
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":""},"usage":{"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"input_tokens":10,"output_tokens":18}}` + "\n\n",
			`data: {"type":"message_stop"}` + "\n\n",
		}

		for _, chunk := range chunks {
			_, _ = fmt.Fprint(w, chunk)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer mockServer.Close()

	client := anthropic.NewClient(
		option.WithAPIKey("test-api-key"),
		option.WithBaseURL(mockServer.URL),
	)

	ctx := context.Background()
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		MaxTokens: 128,
		Model:     anthropic.Model("claude-3-7-sonnet-latest"),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.ContentBlockParamUnion{
				OfText: &anthropic.TextBlockParam{Text: "Hello, Claude"},
			}),
		},
		Temperature: anthropic.Float(0.3),
	})

	for stream.Next() {
		_ = stream.Current()
	}
	if err := stream.Err(); err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		spanStr, _ := json.Marshal(stubs)
		fmt.Println(string(spanStr))
		span := stubs[0][0]

		verifier.Assert(span.Name == "chat stream", "Expected span name to be chat stream, got %s", span.Name)
		operationName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(operationName == "chat", "Expected gen_ai.operation.name to be chat, got %s", operationName)
		requestModel := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(requestModel == "claude-3-7-sonnet-latest", "Expected gen_ai.request.model to be claude-3-7-sonnet-latest, got %s", requestModel)
		genAISpanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
		verifier.Assert(genAISpanKind == "LLM", "Expected gen_ai.span.kind to be LLM, got %s", genAISpanKind)

		isStream := verifier.GetAttribute(span.Attributes, "gen_ai.request.is_stream").AsBool()
		verifier.Assert(isStream, "Expected gen_ai.request.is_stream to be true")

		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 10, "Expected input tokens to be 10, got %d", inputTokens)
		outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
		verifier.Assert(outputTokens == 18, "Expected output tokens to be 18, got %d", outputTokens)
		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 28, "Expected total tokens to be 28, got %d", totalTokens)

		inputMessages := verifier.GetAttribute(span.Attributes, "gen_ai.input.messages").AsString()
		verifier.Assert(len(inputMessages) > 0, "Expected gen_ai.input.messages to be populated")
	}, 1)
}
