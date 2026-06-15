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
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func main() {
	ctx := context.Background()
	
	// Create mock response with non-zero tokens
	response := api.ChatResponse{
		Model:     "llama3:8b",
		CreatedAt: time.Now(),
		Message: api.Message{
			Role:    "assistant",
			Content: "This is a mock chat response with tokens",
		},
		Done: true,
		Metrics: api.Metrics{
			PromptEvalCount:    20, // Input tokens
			EvalCount:          35, // Output tokens
			TotalDuration:      900000000,
			LoadDuration:       90000000,
			PromptEvalDuration: 180000000,
			EvalDuration:       450000000,
		},
	}
	
	server := NewMockOllamaChatServer(response)
	client := NewMockOllamaClient(server)
	defer server.Close()
	
	streamFlag := false
	req := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Test message for chat metrics"},
		},
		Stream: &streamFlag,
	}
	
	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"gen_ai.client.operation.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No gen_ai.client.operation.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count != 1 {
				panic("Expected gen_ai.client.operation.duration count to be 1, got " + strconv.FormatUint(point.DataPoints[0].Count, 10))
			}
			verifier.VerifyGenAIOperationDurationMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "chat", "ollama", "llama3:8b", "llama3:8b")
		},
		"gen_ai.client.token.usage": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No gen_ai.client.token.usage metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[int64])
			
			// Should have 2 data points: one for input, one for output
			if len(point.DataPoints) != 2 {
				panic("Expected 2 data points for gen_ai.client.token.usage, got " + strconv.Itoa(len(point.DataPoints)))
			}
			
			// Verify both data points have correct counts and sums
			inputTokensFound := false
			outputTokensFound := false
			
			for i, dp := range point.DataPoints {
				if dp.Count != 1 {
					panic("Expected gen_ai.client.token.usage data point " + strconv.Itoa(i) + " count to be 1, got " + strconv.FormatUint(dp.Count, 10))
				}
				if dp.Sum <= 0 {
					panic("gen_ai.client.token.usage data point " + strconv.Itoa(i) + " sum is not positive, actually " + strconv.FormatInt(dp.Sum, 10))
				}
				
				// Verify attributes
				verifier.VerifyGenAIOperationDurationMetricsAttributes(dp.Attributes.ToSlice(), "chat", "ollama", "llama3:8b", "llama3:8b")
				
				// Check token type and values
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == "gen_ai.token.type" {
						tokenType := attr.Value.AsString()
						if tokenType == "input" {
							inputTokensFound = true
							if dp.Sum != 20 {
								panic("Expected input tokens sum to be 20, got " + strconv.FormatInt(dp.Sum, 10))
							}
						} else if tokenType == "output" {
							outputTokensFound = true
							if dp.Sum != 35 {
								panic("Expected output tokens sum to be 35, got " + strconv.FormatInt(dp.Sum, 10))
							}
						} else {
							panic("Invalid token type: " + tokenType)
						}
					}
				}
			}
			
			if !inputTokensFound {
				panic("Input tokens (gen_ai.token.type=input) not found in metrics")
			}
			if !outputTokensFound {
				panic("Output tokens (gen_ai.token.type=output) not found in metrics")
			}
		},
	})
}
