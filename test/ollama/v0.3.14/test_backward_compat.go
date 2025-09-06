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
	
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testGenerateStreamFalse(ctx)
	testChatStreamFalse(ctx)
	testGenerateStreamNil(ctx)
	testGenerateStreamTrue(ctx)
}

func testGenerateStreamFalse(ctx context.Context) {
	client, server := SetupMockGenerate(MockGenerateResponse)
	defer server.Close()
	
	streamFalse := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
		Stream: &streamFalse,
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
	
	if chunkCount != 1 {
		panic("Non-streaming should receive exactly 1 chunk")
	}
	
	if finalResponse.Metrics.EvalCount != 25 {
		panic("Unexpected token count")
	}
	
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:8b")
		
		streaming := getAttributeValue(stubs[0][0], "gen_ai.response.streaming")
		if streaming == true {
			panic("Stream=false should not have streaming attribute as true")
		}
	}, 1)
}

func testChatStreamFalse(ctx context.Context) {
	client, server := SetupMockChat(MockChatResponse)
	defer server.Close()
	
	streamFalse := false
	req := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: &streamFalse,
	}
	
	chunkCount := 0
	var finalResponse api.ChatResponse
	
	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		chunkCount++
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	if chunkCount != 1 {
		panic("Non-streaming should receive exactly 1 chunk")
	}
	
	if finalResponse.Metrics.EvalCount != 20 {
		panic("Unexpected token count")
	}
	
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "ollama", "llama3:8b")
		
		streaming := getAttributeValue(stubs[0][0], "gen_ai.response.streaming")
		if streaming == true {
			panic("Stream=false should not have streaming attribute as true")
		}
	}, 1)
}

func testGenerateStreamNil(ctx context.Context) {
	server := NewMockOllamaGenerateStreamServer(MockGenerateStreamChunks)
	client := NewMockOllamaClient(server)
	defer server.Close()
	
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
	}
	
	chunkCount := 0
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	if chunkCount <= 1 {
		panic("Stream=nil should default to streaming with multiple chunks")
	}
	
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:8b")
		VerifyOllamaStreamingAttributes(stubs[0][0])
	}, 1)
}

func testGenerateStreamTrue(ctx context.Context) {
	server := NewMockOllamaGenerateStreamServer(MockGenerateStreamChunks)
	client := NewMockOllamaClient(server)
	defer server.Close()
	
	streamTrue := true
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
		Stream: &streamTrue,
	}
	
	chunkCount := 0
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		return nil
	})
	
	if err != nil {
		panic(err)
	}
	
	if chunkCount != len(MockGenerateStreamChunks) {
		panic("Stream=true should receive all streaming chunks")
	}
	
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "llama3:8b")
		VerifyOllamaStreamingAttributes(stubs[0][0])
	}, 1)
}