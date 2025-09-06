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
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	os.Setenv("OLLAMA_ENABLE_COST_TRACKING", "true")
	os.Setenv("OLLAMA_DEFAULT_CURRENCY", "USD")
	ctx := context.Background()
	testNonStreamingWithCost(ctx)
	testStreamingWithCost(ctx)
	testExpensiveModel(ctx)
}

func testNonStreamingWithCost(ctx context.Context) {
	client, server := SetupMockGenerate(MockGenerateResponse)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
		Stream: &streamFlag,
	}

	var finalResponse api.GenerateResponse
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	if finalResponse.PromptEvalCount != 15 || finalResponse.EvalCount != 25 {
		panic("Unexpected token counts")
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:8b")
		VerifyOllamaCostAttributes(stubs[0][0])
	}, 1)
}

func testStreamingWithCost(ctx context.Context) {
	server := NewMockOllamaGenerateStreamServer(MockGenerateStreamChunks)
	client := NewMockOllamaClient(server)
	defer server.Close()

	streamFlag := true
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
		Stream: &streamFlag,
	}

	chunkCount := 0
	var finalResponse api.GenerateResponse

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	if chunkCount != len(MockGenerateStreamChunks) {
		panic("Unexpected chunk count")
	}

	if finalResponse.EvalCount != 23 {
		panic("Unexpected token count")
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:8b")
		VerifyOllamaStreamingAttributes(stubs[0][0])
		VerifyOllamaCostAttributes(stubs[0][0])
	}, 1)
}

func testExpensiveModel(ctx context.Context) {
	client, server := SetupMockGenerate(ExpensiveGenerateResponse)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:70b",
		Prompt: "Test prompt",
		Stream: &streamFlag,
	}

	var finalResponse api.GenerateResponse
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	if finalResponse.Metrics.PromptEvalCount != 100 || finalResponse.Metrics.EvalCount != 500 {
		panic("Unexpected token counts for expensive model")
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:70b")
		VerifyOllamaCostAttributes(stubs[0][0])
		
		totalCost := getAttributeValue(stubs[0][0], "gen_ai.cost.total_usd")
		if totalCost == nil {
			panic("Cost not calculated for expensive model")
		}
		
		tier := getAttributeValue(stubs[0][0], "gen_ai.cost.model_pricing_tier")
		if tier != "premium" {
			panic("Model not identified as premium tier")
		}
	}, 1)
}