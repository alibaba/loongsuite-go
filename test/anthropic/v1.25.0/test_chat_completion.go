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

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockResponse := `{
"id":"msg_test_123",
"type":"message",
"role":"assistant",
"model":"claude-3-7-sonnet-latest",
"content":[{"type":"text","text":"Hello from anthropic sdk"}],
"stop_reason":"end_turn",
"stop_sequence":"",
"container":{"id":"container_test","expires_at":"2026-01-01T00:00:00Z"},
"usage":{
"cache_creation":{"ephemeral_1h_input_tokens":0,"ephemeral_5m_input_tokens":0},
"cache_creation_input_tokens":0,
"cache_read_input_tokens":0,
"inference_geo":"us",
"input_tokens":12,
"output_tokens":20,
"server_tool_use":{"web_fetch_requests":0,"web_search_requests":0},
"service_tier":"standard"
}
}`
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := anthropic.NewClient(
		option.WithAPIKey("test-api-key"),
		option.WithBaseURL(mockServer.URL),
	)

	ctx := context.Background()
	_, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 128,
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.ContentBlockParamUnion{
				OfText: &anthropic.TextBlockParam{Text: "Hello, Claude"},
			}),
		},
		Temperature: anthropic.Float(0.5),
	})
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		spanStr, _ := json.Marshal(stubs)
		fmt.Println(string(spanStr))
		span := stubs[0][0]

		verifier.Assert(span.Name == "chat", "Expected span name to be chat, got %s", span.Name)
		operationName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
		verifier.Assert(operationName == "chat", "Expected gen_ai.operation.name to be chat, got %s", operationName)
		requestModel := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
		verifier.Assert(requestModel == "claude-3-7-sonnet-latest", "Expected gen_ai.request.model to be claude-3-7-sonnet-latest, got %s", requestModel)
		genAISpanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
		verifier.Assert(genAISpanKind == "LLM", "Expected gen_ai.span.kind to be LLM, got %s", genAISpanKind)

		maxTokens := verifier.GetAttribute(span.Attributes, "gen_ai.request.max_tokens").AsInt64()
		verifier.Assert(maxTokens == 128, "Expected max_tokens to be 128, got %d", maxTokens)

		inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
		verifier.Assert(inputTokens == 12, "Expected input tokens to be 12, got %d", inputTokens)
		outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
		verifier.Assert(outputTokens == 20, "Expected output tokens to be 20, got %d", outputTokens)
		totalTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.total_tokens").AsInt64()
		verifier.Assert(totalTokens == 32, "Expected total tokens to be 32, got %d", totalTokens)

		finishReasons := verifier.GetAttribute(span.Attributes, "gen_ai.response.finish_reasons").AsStringSlice()
		verifier.Assert(len(finishReasons) == 1 && finishReasons[0] == "end_turn", "Expected finish reason to be [end_turn], got %v", finishReasons)
	}, 1)
}
