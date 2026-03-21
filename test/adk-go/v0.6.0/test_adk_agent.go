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
	"net/http"
	"net/http/httptest"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {"role": "model", "parts": [{"text": "Hello from mock Gemini!"}]},
				"finishReason": "STOP"
			}],
			"usageMetadata": {
				"promptTokenCount": 10,
				"candidatesTokenCount": 5,
				"totalTokenCount": 15
			},
			"modelVersion": "gemini-2.5-flash-001"
		}`))
	}))
	defer mockServer.Close()

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey:  "fake-api-key",
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			BaseURL: mockServer.URL + "/",
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create model: %v", err))
	}

	myAgent, err := llmagent.New(llmagent.Config{
		Name:        "test-agent",
		Description: "A test agent",
		Model:       model,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create agent: %v", err))
	}

	sessSvc := session.InMemoryService()
	sessResp, err := sessSvc.Create(ctx, &session.CreateRequest{
		AppName: "adk-go-test",
		UserID:  "test-user",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create session: %v", err))
	}

	r, err := runner.New(runner.Config{
		AppName:        "adk-go-test",
		Agent:          myAgent,
		SessionService: sessSvc,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create runner: %v", err))
	}

	msg := genai.NewContentFromText("Hello, say hi!", "user")
	for ev, err := range r.Run(ctx, "test-user", sessResp.Session.ID(), msg, agent.RunConfig{}) {
		if err != nil {
			panic(fmt.Sprintf("runner error: %v", err))
		}
		if ev != nil {
			fmt.Printf("Event: author=%s\n", ev.Author)
		}
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.Assert(len(stubs) >= 1, "Expected at least 1 trace, got %d", len(stubs))

		foundWorkflow := false
		foundLLM := false
		for _, spans := range stubs {
			for _, span := range spans {
				system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
				if system != "google_adk" {
					continue
				}
				opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()

				if opName == "invoke_agent" && !foundWorkflow {
					foundWorkflow = true
					spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
					verifier.Assert(spanKind == "workflow", "Expected gen_ai.span.kind=workflow, got %s", spanKind)
					userID := verifier.GetAttribute(span.Attributes, "gen_ai.other_input.user_id").AsString()
					verifier.Assert(userID == "test-user", "Expected user_id=test-user, got %s", userID)
					verifier.Assert(span.SpanKind == trace.SpanKindClient, "Expected client span, got %d", span.SpanKind)
				}

				if opName == "chat" && !foundLLM {
					foundLLM = true
					modelName := verifier.GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
					verifier.Assert(modelName != "", "Expected non-empty gen_ai.request.model")
					spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
					verifier.Assert(spanKind == "generation", "Expected gen_ai.span.kind=generation, got %s", spanKind)
					verifier.Assert(span.SpanKind == trace.SpanKindClient, "Expected client span, got %d", span.SpanKind)

					inputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.input_tokens").AsInt64()
					verifier.Assert(inputTokens == 10, "Expected input_tokens=10, got %d", inputTokens)
					outputTokens := verifier.GetAttribute(span.Attributes, "gen_ai.usage.output_tokens").AsInt64()
					verifier.Assert(outputTokens == 5, "Expected output_tokens=5, got %d", outputTokens)
					finishReasons := verifier.GetAttribute(span.Attributes, "gen_ai.response.finish_reasons").AsStringSlice()
					verifier.Assert(len(finishReasons) == 1 && finishReasons[0] == "STOP", "Expected finish_reasons=[STOP], got %v", finishReasons)
					responseModel := verifier.GetAttribute(span.Attributes, "gen_ai.response.model").AsString()
					verifier.Assert(responseModel == "gemini-2.5-flash-001", "Expected response.model=gemini-2.5-flash-001, got %s", responseModel)
				}
			}
		}
		verifier.Assert(foundWorkflow, "Expected to find workflow span (invoke_agent)")
		verifier.Assert(foundLLM, "Expected to find LLM span (chat) with response data")
	}, 1)
}
