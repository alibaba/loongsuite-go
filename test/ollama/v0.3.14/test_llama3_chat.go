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
	"net/http/httptest"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	response := api.ChatResponse{
		Model:     "llama3:70b",
		CreatedAt: time.Now(),
		Message: api.Message{
			Role:    "assistant",
			Content: "This is a response from the large llama3:70b model",
		},
		Done:            true,
		Metrics: api.Metrics{
			PromptEvalCount: 50,
			EvalCount:       100,
			TotalDuration:   3000000000,
			LoadDuration:    300000000,
			PromptEvalDuration: 150000000,
			EvalDuration:    500000000,
		},
	}
	server := NewMockOllamaChatServer(response)
	client := NewMockOllamaClient(server)
	defer server.Close()
	streamFlag := false
	req := &api.ChatRequest{
		Model: "llama3:70b",
		Messages: []api.Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
		Stream: &streamFlag,
	}
	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "ollama", "llama3:70b")
		for _, attr := range stubs[0][0].Attributes {
			if string(attr.Key) == "gen_ai.cost.model_pricing_tier" {
				if attr.Value.AsString() != "premium" {
					panic("llama3:70b should be marked as premium tier")
				}
			}
		}
	}, 1)
}

func NewMockLlama3Server() *httptest.Server {
	response := api.ChatResponse{
		Model:     "llama3:70b",
		CreatedAt: time.Now(),
		Message: api.Message{
			Role:    "assistant",
			Content: "This is a response from the large llama3:70b model",
		},
		Done:            true,
		Metrics: api.Metrics{
			PromptEvalCount: 50,
			EvalCount:       100,
			TotalDuration:   3000000000,
		},
	}
	return NewMockOllamaChatServer(response)
}