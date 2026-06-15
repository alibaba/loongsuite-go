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
	
	// Create streaming responses with non-zero tokens
	responses := []api.GenerateResponse{
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "This ", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "is ", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "a ", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "streaming ", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "response", Done: true, 
			Metrics: api.Metrics{
				PromptEvalCount:    30, // Input tokens
				EvalCount:          45, // Output tokens
				TotalDuration:      800000000,
				LoadDuration:       80000000,
				PromptEvalDuration: 200000000,
				EvalDuration:       400000000,
			},
		},
	}
	
	server := NewMockOllamaGenerateStreamServer(responses)
	client := NewMockOllamaClient(server)
	defer server.Close()
	
	streamFlag := true
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test streaming prompt for metrics",
		Stream: &streamFlag,
	}
	
	chunkCount := 0
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	if chunkCount < 2 {
		panic("Expected at least 2 chunks for streaming, got " + strconv.Itoa(chunkCount))
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
			verifier.VerifyGenAIOperationDurationMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "generate", "ollama", "llama3:8b", "llama3:8b")
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
				verifier.VerifyGenAIOperationDurationMetricsAttributes(dp.Attributes.ToSlice(), "generate", "ollama", "llama3:8b", "llama3:8b")
				
				// Check token type and values
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == "gen_ai.token.type" {
						tokenType := attr.Value.AsString()
						if tokenType == "input" {
							inputTokensFound = true
							if dp.Sum != 30 {
								panic("Expected input tokens sum to be 30, got " + strconv.FormatInt(dp.Sum, 10))
							}
						} else if tokenType == "output" {
							outputTokensFound = true
							if dp.Sum != 45 {
								panic("Expected output tokens sum to be 45, got " + strconv.FormatInt(dp.Sum, 10))
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
		"gen_ai.server.time_to_first_token": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No gen_ai.server.time_to_first_token metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count != 1 {
				panic("Expected gen_ai.server.time_to_first_token count to be 1, got " + strconv.FormatUint(point.DataPoints[0].Count, 10))
			}
			if point.DataPoints[0].Sum <= 0 {
				panic("gen_ai.server.time_to_first_token sum should be positive")
			}
			verifier.VerifyGenAIOperationDurationMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "generate", "ollama", "llama3:8b", "llama3:8b")
		},
	})
}
