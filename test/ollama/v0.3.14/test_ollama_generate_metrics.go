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
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	
	// Create mock response with non-zero tokens
	response := api.GenerateResponse{
		Model:     "llama3:8b",
		CreatedAt: time.Now(),
		Response:  "This is a mock response with tokens",
		Done:      true,
		Metrics: api.Metrics{
			PromptEvalCount:    25, // Input tokens
			EvalCount:          50, // Output tokens
			TotalDuration:      1000000000,
			LoadDuration:       100000000,
			PromptEvalDuration: 250000000,
			EvalDuration:       500000000,
		},
	}
	
	server := NewMockOllamaGenerateServer(response)
	client := NewMockOllamaClient(server)
	defer server.Close()
	
	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt for metrics",
		Stream: &streamFlag,
	}
	
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
	
	// Verify traces first to ensure span attributes have correct token values
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) == 0 || len(stubs[0]) == 0 {
			panic("No traces received")
		}
		span := stubs[0][0]
		
		// Verify token values in span attributes
		var inputTokens, outputTokens int64
		var hasInputTokens, hasOutputTokens bool
		for _, attr := range span.Attributes {
			if attr.Key == "gen_ai.usage.input_tokens" {
				inputTokens = attr.Value.AsInt64()
				hasInputTokens = true
			}
			if attr.Key == "gen_ai.usage.output_tokens" {
				outputTokens = attr.Value.AsInt64()
				hasOutputTokens = true
			}
		}
		
		if !hasInputTokens {
			panic("gen_ai.usage.input_tokens not found in span attributes")
		}
		if !hasOutputTokens {
			panic("gen_ai.usage.output_tokens not found in span attributes")
		}
		if inputTokens != 25 {
			panic("Expected input tokens to be 25 in span attributes, got " + strconv.FormatInt(inputTokens, 10))
		}
		if outputTokens != 50 {
			panic("Expected output tokens to be 50 in span attributes, got " + strconv.FormatInt(outputTokens, 10))
		}
	}, 1)
	
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
			for i, dp := range point.DataPoints {
				if dp.Count != 1 {
					panic("Expected gen_ai.client.token.usage data point " + strconv.Itoa(i) + " count to be 1, got " + strconv.FormatUint(dp.Count, 10))
				}
				if dp.Sum <= 0 {
					panic("gen_ai.client.token.usage data point " + strconv.Itoa(i) + " sum is not positive, actually " + strconv.FormatInt(dp.Sum, 10))
				}
				
				// Verify attributes
				verifier.VerifyGenAIOperationDurationMetricsAttributes(dp.Attributes.ToSlice(), "generate", "ollama", "llama3:8b", "llama3:8b")
				
				// Check token type attribute exists
				hasTokenType := false
				for _, attr := range dp.Attributes.ToSlice() {
					if attr.Key == "gen_ai.token.type" {
						hasTokenType = true
						tokenType := attr.Value.AsString()
						if tokenType == "input" {
							if dp.Sum != 25 {
								panic("Expected input tokens sum to be 25, got " + strconv.FormatInt(dp.Sum, 10))
							}
						} else if tokenType == "output" {
							if dp.Sum != 50 {
								panic("Expected output tokens sum to be 50, got " + strconv.FormatInt(dp.Sum, 10))
							}
						} else {
							panic("Invalid token type: " + tokenType)
						}
					}
				}
				if !hasTokenType {
					panic("gen_ai.token.type attribute not found in data point " + strconv.Itoa(i))
				}
			}
		},
	})
}
