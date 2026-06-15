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
	response := api.GenerateResponse{
		Model:           "tinyllama",
		CreatedAt:       time.Now(),
		Response:        "This is a response from tinyllama model",
		Done:            true,
		Metrics: api.Metrics{
			PromptEvalCount: 10,
			EvalCount:       15,
			TotalDuration:   500000000,
			LoadDuration:    50000000,
			PromptEvalDuration: 25000000,
			EvalDuration:    75000000,
		},
	}
	server := NewMockOllamaGenerateServer(response)
	client := NewMockOllamaClient(server)
	defer server.Close()
	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Hello",
		Stream: &streamFlag,
	}
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "generate", "ollama", "tinyllama")
	}, 1)
}

func NewMockTinyLlamaServer() *httptest.Server {
	response := api.GenerateResponse{
		Model:           "tinyllama",
		CreatedAt:       time.Now(),
		Response:        "This is a response from tinyllama model",
		Done:            true,
		Metrics: api.Metrics{
			PromptEvalCount: 10,
			EvalCount:       15,
			TotalDuration:   500000000,
		},
	}
	return NewMockOllamaGenerateServer(response)
}