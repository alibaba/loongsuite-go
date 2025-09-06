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
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	os.Setenv("OLLAMA_ENABLE_COST_TRACKING", "true")
	os.Setenv("OLLAMA_DEFAULT_CURRENCY", "USD")
	os.Setenv("OLLAMA_BUDGET_TOTAL", "0.0001")
	os.Setenv("OLLAMA_BUDGET_PERIOD", "hourly")

	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/generate" {
			responses := []api.GenerateResponse{
				{
					Model:     "llama3:8b",
					CreatedAt: time.Now(),
					Response:  "First response",
					Done:      true,
					Metrics: api.Metrics{
						TotalDuration:      500000000,
						LoadDuration:       50000000,
						PromptEvalCount:    10,
						PromptEvalDuration: 100000000,
						EvalCount:          20,
						EvalDuration:       200000000,
					},
				},
				{
					Model:     "llama3:8b",
					CreatedAt: time.Now(),
					Response:  "Second response with more tokens",
					Done:      true,
					Metrics: api.Metrics{
						TotalDuration:      800000000,
						LoadDuration:       80000000,
						PromptEvalCount:    25,
						PromptEvalDuration: 150000000,
						EvalCount:          50,
						EvalDuration:       400000000,
					},
				},
				{
					Model:     "llama3:8b",
					CreatedAt: time.Now(),
					Response:  "Third response exceeding budget",
					Done:      true,
					Metrics: api.Metrics{
						TotalDuration:      1200000000,
						LoadDuration:       100000000,
						PromptEvalCount:    100,
						PromptEvalDuration: 300000000,
						EvalCount:          200,
						EvalDuration:       800000000,
					},
				},
			}

			resp := responses[0]
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewMockOllamaClient(server)

	testProgressiveBudgetConsumption(ctx, client)

	testBudgetThresholds(ctx, client)

	testAnomalyDetection(ctx, client)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs[0]) < 1 {
			panic("Expected at least 1 span for budget tracking")
		}

		for _, span := range stubs[0] {
			verifier.VerifyLLMAttributes(span, "generate", "ollama", "llama3:8b")
			
			hasbudgetAttr := false
			for _, attr := range span.Attributes {
				switch attr.Key {
				case "gen_ai.cost.total_usd", "gen_ai.cost.input_tokens_usd", "gen_ai.cost.output_tokens_usd":
					hasbudgetAttr = true
				case "gen_ai.budget.remaining_usd", "gen_ai.budget.usage_percentage":
					hasbudgetAttr = true
				}
			}
			
			if hasbudgetAttr {
				return
			}
		}
	}, 1)
}

func testProgressiveBudgetConsumption(ctx context.Context, client *api.Client) {
	for i := 0; i < 3; i++ {
		req := &api.GenerateRequest{
			Model:  "llama3:8b",
			Prompt: "Test prompt for budget tracking",
			Stream: new(bool),
		}
		*req.Stream = false

		err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
			return nil
		})
		if err != nil {
			return
		}
	}
}

func testBudgetThresholds(ctx context.Context, client *api.Client) {
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Long prompt to test budget thresholds with many tokens",
		Stream: new(bool),
	}
	*req.Stream = false

	client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
}

func testAnomalyDetection(ctx context.Context, client *api.Client) {
	normalPrompts := []string{"Hi", "Hello", "Test"}
	for _, prompt := range normalPrompts {
		req := &api.GenerateRequest{
			Model:  "llama3:8b",
			Prompt: prompt,
			Stream: new(bool),
		}
		*req.Stream = false

		client.Generate(ctx, req, func(resp api.GenerateResponse) error {
			return nil
		})
	}

	anomalousReq := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Write a very long story about AI and include many details. This is an anomalously long prompt that should trigger anomaly detection in the budget tracking system.",
		Stream: new(bool),
	}
	*anomalousReq.Stream = false

	client.Generate(ctx, anomalousReq, func(resp api.GenerateResponse) error {
		return nil
	})
}