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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func NewMockOllamaGenerateServer(response api.GenerateResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

func NewMockOllamaGenerateStreamServer(chunks []api.GenerateResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		
		encoder := json.NewEncoder(w)
		for _, chunk := range chunks {
			if err := encoder.Encode(chunk); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
}

func NewMockOllamaChatServer(response api.ChatResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

func NewMockOllamaChatStreamServer(chunks []api.ChatResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		
		encoder := json.NewEncoder(w)
		for _, chunk := range chunks {
			if err := encoder.Encode(chunk); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
}

func NewMockOllamaErrorServer(statusCode int, errorMsg string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errorMsg,
		})
	}))
}

func NewMockOllamaClient(server *httptest.Server) *api.Client {
	u, _ := url.Parse(server.URL)
	return api.NewClient(u, http.DefaultClient)
}

func SetupMockGenerate(response api.GenerateResponse) (*api.Client, *httptest.Server) {
	server := NewMockOllamaGenerateServer(response)
	client := NewMockOllamaClient(server)
	return client, server
}

func SetupMockChat(response api.ChatResponse) (*api.Client, *httptest.Server) {
	server := NewMockOllamaChatServer(response)
	client := NewMockOllamaClient(server)
	return client, server
}

func NewMockOllamaGenerateForInvoke(ctx context.Context) (*api.Client, *httptest.Server) {
	return SetupMockGenerate(MockGenerateResponse)
}

func NewMockOllamaGenerateForStream(ctx context.Context) (*api.Client, *httptest.Server) {
	server := NewMockOllamaGenerateStreamServer(MockGenerateStreamChunks)
	client := NewMockOllamaClient(server)
	return client, server
}

func NewMockOllamaChatForInvoke(ctx context.Context) (*api.Client, *httptest.Server) {
	return SetupMockChat(MockChatResponse)
}

func NewMockOllamaChatForStream(ctx context.Context) (*api.Client, *httptest.Server) {
	server := NewMockOllamaChatStreamServer(MockChatStreamChunks)
	client := NewMockOllamaClient(server)
	return client, server
}

func NewMockOllamaGenerateWithTokens(ctx context.Context, promptTokens, completionTokens int) (*api.Client, *httptest.Server) {
	response := api.GenerateResponse{
		Model:     "llama3.2:1b",
		CreatedAt: time.Now(),
		Response:  "Response with specific token counts",
		Done:      true,
		Metrics: api.Metrics{
			TotalDuration:      1000000000,
			LoadDuration:       100000000,
			PromptEvalCount:    promptTokens,
			PromptEvalDuration: 200000000,
			EvalCount:          completionTokens,
			EvalDuration:       400000000,
		},
	}
	server := NewMockOllamaGenerateServer(response)
	client := NewMockOllamaClient(server)
	return client, server
}

var (
	MockGenerateResponse = api.GenerateResponse{
		Model:           "llama3:8b",
		CreatedAt:       time.Now(),
		Response:        "This is a mock generated response",
		Done:            true,
		Context:         []int{1, 2, 3, 4, 5},
		Metrics: api.Metrics{
			TotalDuration:   1000000000, 
			LoadDuration:    100000000,  
			PromptEvalCount: 15,
			PromptEvalDuration: 50000000,
			EvalCount:       25,
			EvalDuration:    150000000, 
		},
	}
	
	MockGenerateStreamChunks = []api.GenerateResponse{
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "Hello", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: " from", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: " mock", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: " streaming", Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Response: "!", Done: true, 
			Metrics: api.Metrics{
				PromptEvalCount: 10, 
				EvalCount: 23, 
				TotalDuration: 500000000,
				LoadDuration: 50000000, 
				PromptEvalDuration: 30000000, 
				EvalDuration: 100000000,
			},
		},
	}
	
	MockChatResponse = api.ChatResponse{
		Model:     "llama3:8b",
		CreatedAt: time.Now(),
		Message: api.Message{
			Role:    "assistant",
			Content: "This is a mock chat response",
		},
		Done:            true,
		Metrics: api.Metrics{
			TotalDuration:   800000000, 
			LoadDuration:    80000000, 
			PromptEvalCount: 12,
			PromptEvalDuration: 40000000, 
			EvalCount:       20,
			EvalDuration:    120000000,
		},
	}
	
	MockChatStreamChunks = []api.ChatResponse{
		{Model: "llama3:8b", CreatedAt: time.Now(), Message: api.Message{Role: "assistant", Content: "Hi"}, Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Message: api.Message{Role: "assistant", Content: " there"}, Done: false},
		{Model: "llama3:8b", CreatedAt: time.Now(), Message: api.Message{Role: "assistant", Content: "!"}, Done: true,
			Metrics: api.Metrics{
				PromptEvalCount: 8, 
				EvalCount: 15, 
				TotalDuration: 400000000,
				LoadDuration: 40000000, 
				PromptEvalDuration: 20000000, 
				EvalDuration: 80000000,
			},
		},
	}
	
	ExpensiveGenerateResponse = api.GenerateResponse{
		Model:           "llama3:70b",
		Response:        "Response from expensive model",
		Done:            true,
		Metrics: api.Metrics{
			PromptEvalCount: 100,
			EvalCount:       500,
			TotalDuration:   5000000000,
		},
	}
)

func VerifyOllamaAttributes(span tracetest.SpanStub, operation, model string) {
	expectedSpanName := fmt.Sprintf("ollama.%s", operation)
	if span.Name != expectedSpanName {
		panic(fmt.Sprintf("Expected span name %s, got %s", expectedSpanName, span.Name))
	}
	
	verifier.VerifyLLMAttributes(span, operation, "ollama", model)
	
	requiredAttrs := map[string]bool{
		"gen_ai.system":         false,
		"gen_ai.request.model":  false,
		"gen_ai.operation.name": false,
	}
	
	for _, attr := range span.Attributes {
		key := string(attr.Key)
		if _, ok := requiredAttrs[key]; ok {
			requiredAttrs[key] = true
		}
	}
	
	for attr, found := range requiredAttrs {
		if !found {
			panic(fmt.Sprintf("Required attribute %s not found", attr))
		}
	}
	
	inputTokens := getAttributeValue(span, "gen_ai.usage.input_tokens")
	outputTokens := getAttributeValue(span, "gen_ai.usage.output_tokens")
	
	if inputTokens == nil || outputTokens == nil {
		panic(fmt.Sprintf("Token counts not found - input: %v, output: %v", inputTokens, outputTokens))
	}
	
	if inputTokensInt, ok := inputTokens.(int64); ok && inputTokensInt <= 0 {
		panic(fmt.Sprintf("Invalid input token count: %d", inputTokensInt))
	}
	if outputTokensInt, ok := outputTokens.(int64); ok && outputTokensInt <= 0 {
		panic(fmt.Sprintf("Invalid output token count: %d", outputTokensInt))
	}
	
	totalDuration := getAttributeValue(span, "gen_ai.response.total_duration_ms")
	if totalDuration != nil {
		if durationMs, ok := totalDuration.(float64); ok && durationMs <= 0 {
			panic(fmt.Sprintf("Invalid total duration: %f", durationMs))
		}
	}
}

func VerifyOllamaStreamingAttributes(span tracetest.SpanStub) {
	streaming := getAttributeValue(span, "gen_ai.response.streaming")
	if streaming != true {
		panic(fmt.Sprintf("Expected streaming=true, got %v", streaming))
	}
}

func VerifyOllamaCostAttributes(span tracetest.SpanStub) {
	requiredAttrs := []string{
		"gen_ai.cost.total_usd",
		"gen_ai.cost.input_tokens_usd",
		"gen_ai.cost.output_tokens_usd",
		"gen_ai.cost.currency",
		"gen_ai.cost.model_pricing_tier",
	}
	
	for _, attr := range requiredAttrs {
		if getAttributeValue(span, attr) == nil {
			panic(fmt.Sprintf("Missing cost attribute: %s", attr))
		}
	}
}

func VerifyOllamaBudgetAttributes(span tracetest.SpanStub) {
	requiredAttrs := []string{
		"gen_ai.budget.status",
		"gen_ai.budget.usage_percentage",
		"gen_ai.budget.remaining_usd",
		"gen_ai.budget.threshold_exceeded",
	}
	
	for _, attr := range requiredAttrs {
		if getAttributeValue(span, attr) == nil {
			panic(fmt.Sprintf("Missing budget attribute: %s", attr))
		}
	}
}

func VerifyOllamaErrorSpan(span tracetest.SpanStub) {
	if span.Status.Code != codes.Error {
		panic(fmt.Sprintf("Expected error status, got %v", span.Status.Code))
	}
	
	if span.Status.Description == "" {
		panic("Error description not set in span")
	}
}

func getAttributeValue(span tracetest.SpanStub, key string) interface{} {
	for _, attr := range span.Attributes {
		if string(attr.Key) == key {
			return attr.Value.AsInterface()
		}
	}
	return nil
}

func CreateSimpleTestClient(ctx context.Context, mockType string) (*api.Client, *httptest.Server) {
	switch mockType {
	case "generate":
		return SetupMockGenerate(MockGenerateResponse)
	case "generate-stream":
		server := NewMockOllamaGenerateStreamServer(MockGenerateStreamChunks)
		return NewMockOllamaClient(server), server
	case "chat":
		return SetupMockChat(MockChatResponse)
	case "chat-stream":
		server := NewMockOllamaChatStreamServer(MockChatStreamChunks)
		return NewMockOllamaClient(server), server
	case "error":
		server := NewMockOllamaErrorServer(404, "model not found")
		return NewMockOllamaClient(server), server
	case "expensive":
		return SetupMockGenerate(ExpensiveGenerateResponse)
	default:
		panic("Unknown mock type: " + mockType)
	}
}
const testModel = "llama3.2:latest"

func checkSpanAttributes(requiredAttributes []string) []string {
	var missing []string
	for _, attr := range requiredAttributes {
		missing = append(missing, attr)
	}
	return missing
}

func isServerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	switch e := err.(type) {
	case *api.StatusError:
		return e.StatusCode == http.StatusServiceUnavailable
	default:
		return false
	}
}
