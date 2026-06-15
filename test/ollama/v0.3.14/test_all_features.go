// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()

	testBasicGenerate(ctx)
	testBasicChat(ctx)

	testStreamingGenerate(ctx)
	testStreamingChat(ctx)

	testCostTracking(ctx)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) < 5 {
			panic("Expected at least 5 traces for all features test")
		}

		hasChat := false
		hasGenerate := false
		for _, trace := range stubs {
			for _, span := range trace {
				for _, attr := range span.Attributes {
					switch attr.Key {
					case "gen_ai.operation.name":
						if attr.Value.AsString() == "chat" {
							hasChat = true
						} else if attr.Value.AsString() == "generate" {
							hasGenerate = true
						}
					}
				}
			}
		}

		if !hasChat || !hasGenerate {
			panic("Not all features were properly tested")
		}
	}, 5)
}

func testBasicGenerate(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test basic generate",
		Stream: new(bool),
	}
	*req.Stream = false

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func testBasicChat(ctx context.Context) {
	client, server := NewMockOllamaChatForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Test basic chat"},
		},
		Stream: &streamFlag,
	}

	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func testStreamingGenerate(ctx context.Context) {
	client, server := NewMockOllamaGenerateForStream(ctx)
	defer server.Close()

	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test streaming generate",
		Stream: new(bool),
	}
	*req.Stream = true

	chunkCount := 0
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		return nil
	})
	if err != nil {
		panic(err)
	}

	if chunkCount < 2 {
		panic("Expected at least 2 chunks for streaming")
	}
}

func testStreamingChat(ctx context.Context) {
	client, server := NewMockOllamaChatForStream(ctx)
	defer server.Close()

	streamFlag := true
	req := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Test streaming chat"},
		},
		Stream: &streamFlag,
	}

	chunkCount := 0
	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		chunkCount++
		return nil
	})
	if err != nil {
		panic(err)
	}

	if chunkCount < 2 {
		panic("Expected at least 2 chunks for streaming")
	}
}

func testCostTracking(ctx context.Context) {
	client, server := NewMockOllamaGenerateWithTokens(ctx, 50, 100)
	defer server.Close()

	req := &api.GenerateRequest{
		Model:  "llama3.2:1b",
		Prompt: "Calculate costs",
		Stream: new(bool),
	}
	*req.Stream = false

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

	if finalResponse.Metrics.PromptEvalCount != 50 || finalResponse.Metrics.EvalCount != 100 {
		panic("Token counts don't match expected values for cost calculation")
	}
}
