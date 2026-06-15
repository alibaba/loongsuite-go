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
	"fmt"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testOptionsExtractionGenerate(ctx)
	testOptionsExtractionChat(ctx)
	testOptionsExtractionStream(ctx)
	testOptionsWithDifferentTypes(ctx)
}

func testOptionsExtractionGenerate(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test prompt",
		Stream: &streamFlag,
		Options: map[string]interface{}{
			"temperature":       0.7,
			"num_predict":       1024,
			"top_k":             40.0,
			"top_p":             0.9,
			"frequency_penalty": 0.5,
			"presence_penalty":  0.3,
			"stop":              []string{"\n\n", "END"},
			"seed":              42,
		},
	}

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		verifier.VerifyLLMAttributes(span, "generate", "ollama", "llama3:8b")

		verifyOptionsAttributes(span, map[string]interface{}{
			"gen_ai.request.temperature":        0.7,
			"gen_ai.request.max_tokens":         int64(1024),
			"gen_ai.request.top_k":              40.0,
			"gen_ai.request.top_p":              0.9,
			"gen_ai.request.frequency_penalty":  0.5,
			"gen_ai.request.presence_penalty":   0.3,
			"gen_ai.request.stop_sequences":     []string{"\n\n", "END"},
			"gen_ai.request.seed":               int64(42),
		})
	}, 1)
}

func testOptionsExtractionChat(ctx context.Context) {
	client, server := NewMockOllamaChatForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: &streamFlag,
		Options: map[string]interface{}{
			"temperature": 0.8,
			"num_predict": 512,
			"top_k":       50.0,
			"top_p":       0.95,
			"seed":        123,
		},
	}

	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		verifier.VerifyLLMAttributes(span, "chat", "ollama", "llama3:8b")

		verifyOptionsAttributes(span, map[string]interface{}{
			"gen_ai.request.temperature": 0.8,
			"gen_ai.request.max_tokens":  int64(512),
			"gen_ai.request.top_k":       50.0,
			"gen_ai.request.top_p":       0.95,
			"gen_ai.request.seed":        int64(123),
		})
	}, 1)
}

func testOptionsExtractionStream(ctx context.Context) {
	client, server := NewMockOllamaGenerateForStream(ctx)
	defer server.Close()

	streamFlag := true
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Streaming test",
		Stream: &streamFlag,
		Options: map[string]interface{}{
			"temperature": 0.5,
			"top_k":       30.0,
			"top_p":       0.85,
		},
	}

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		verifier.VerifyLLMAttributes(span, "generate", "ollama", "llama3:8b")
		VerifyOllamaStreamingAttributes(span)

		verifyOptionsAttributes(span, map[string]interface{}{
			"gen_ai.request.temperature": 0.5,
			"gen_ai.request.top_k":       30.0,
			"gen_ai.request.top_p":       0.85,
		})
	}, 1)
}

func testOptionsWithDifferentTypes(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test different types",
		Stream: &streamFlag,
		Options: map[string]interface{}{
			"temperature":       float32(0.6),
			"num_predict":       int(2048),
			"top_k":             int64(45),
			"frequency_penalty": float32(0.4),
			"stop":              []interface{}{"STOP", "END", "\n\n\n"},
		},
	}

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		verifier.VerifyLLMAttributes(span, "generate", "ollama", "llama3:8b")

		temp := getAttributeValue(span, "gen_ai.request.temperature")
		if temp == nil {
			panic("Temperature should be extracted from float32")
		}

		maxTokens := getAttributeValue(span, "gen_ai.request.max_tokens")
		if maxTokens == nil {
			panic("Max tokens should be extracted from int")
		}

		topK := getAttributeValue(span, "gen_ai.request.top_k")
		if topK == nil {
			panic("Top K should be extracted from int64")
		}

		stopSeqs := getAttributeValue(span, "gen_ai.request.stop_sequences")
		if stopSeqs == nil {
			panic("Stop sequences should be extracted from []interface{}")
		}
		stopSeqsSlice, ok := stopSeqs.([]string)
		if !ok || len(stopSeqsSlice) != 3 {
			panic(fmt.Sprintf("Stop sequences should be []string with 3 elements, got %v", stopSeqs))
		}
	}, 1)
}

func verifyOptionsAttributes(span tracetest.SpanStub, expectedAttrs map[string]interface{}) {
	for key, expectedValue := range expectedAttrs {
		actualValue := getAttributeValue(span, key)
		if actualValue == nil {
			panic(fmt.Sprintf("Expected attribute %s not found", key))
		}

		switch expected := expectedValue.(type) {
		case float64:
			actual, ok := actualValue.(float64)
			if !ok {
				panic(fmt.Sprintf("Attribute %s: expected float64, got %T", key, actualValue))
			}
			if actual != expected {
				panic(fmt.Sprintf("Attribute %s: expected %v, got %v", key, expected, actual))
			}
		case int64:
			actual, ok := actualValue.(int64)
			if !ok {
				panic(fmt.Sprintf("Attribute %s: expected int64, got %T", key, actualValue))
			}
			if actual != expected {
				panic(fmt.Sprintf("Attribute %s: expected %v, got %v", key, expected, actual))
			}
		case []string:
			actual, ok := actualValue.([]string)
			if !ok {
				panic(fmt.Sprintf("Attribute %s: expected []string, got %T", key, actualValue))
			}
			if len(actual) != len(expected) {
				panic(fmt.Sprintf("Attribute %s: expected length %d, got %d", key, len(expected), len(actual)))
			}
			for i := range expected {
				if actual[i] != expected[i] {
					panic(fmt.Sprintf("Attribute %s[%d]: expected %s, got %s", key, i, expected[i], actual[i]))
				}
			}
		default:
			panic(fmt.Sprintf("Unsupported expected type %T for attribute %s", expectedValue, key))
		}
	}
}
