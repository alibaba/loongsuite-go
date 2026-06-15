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
	testStandardAttributesOnly(ctx)
	testNoDeletedAttributes(ctx)
}

func testStandardAttributesOnly(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test standard attributes",
		Stream: &streamFlag,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 1024,
			"top_k":       40.0,
			"top_p":       0.9,
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

		requiredAttrs := []string{
			"gen_ai.system",
			"gen_ai.request.model",
			"gen_ai.operation.name",
			"gen_ai.response.model",
			"gen_ai.response.finish_reasons",
			"gen_ai.usage.input_tokens",
			"gen_ai.usage.output_tokens",
			"server.address",
		}

		for _, attr := range requiredAttrs {
			value := getAttributeValue(span, attr)
			if value == nil {
				panic(fmt.Sprintf("Required standard attribute %s not found", attr))
			}
		}
	}, 1)
}

func testNoDeletedAttributes(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test no deleted attributes",
		Stream: &streamFlag,
	}

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]

		deletedAttrs := []string{
			"gen_ai.request.encoding_formats",
			"gen_ai.cost.total_usd",
			"gen_ai.cost.input_tokens_usd",
			"gen_ai.cost.output_tokens_usd",
			"gen_ai.cost.currency",
			"gen_ai.cost.model_pricing_tier",
			"gen_ai.budget.status",
			"gen_ai.budget.usage_percentage",
			"gen_ai.budget.remaining_usd",
			"gen_ai.budget.threshold_exceeded",
			"gen_ai.slo.latency_threshold_ms",
			"gen_ai.slo.latency_violation",
		}

		for _, attr := range deletedAttrs {
			value := getAttributeValue(span, attr)
			if value != nil {
				panic(fmt.Sprintf("Deleted attribute %s should not be present, but found with value: %v", attr, value))
			}
		}
	}, 1)
}
