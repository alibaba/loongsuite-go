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
	"strings"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testServerAddressExtraction(ctx)
	testServerAddressWithDifferentPorts(ctx)
}

func testServerAddressExtraction(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test server address",
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
		verifier.VerifyLLMAttributes(span, "generate", "ollama", "llama3:8b")

		serverAddress := getAttributeValue(span, "server.address")
		if serverAddress == nil {
			panic("server.address attribute should be present")
		}

		serverAddrStr, ok := serverAddress.(string)
		if !ok {
			panic(fmt.Sprintf("server.address should be string, got %T", serverAddress))
		}

		if serverAddrStr == "" {
			panic("server.address should not be empty")
		}

		if !strings.Contains(serverAddrStr, "127.0.0.1") && !strings.Contains(serverAddrStr, "localhost") {
			panic(fmt.Sprintf("server.address should contain localhost or 127.0.0.1, got %s", serverAddrStr))
		}

		if !strings.Contains(serverAddrStr, ":") {
			panic(fmt.Sprintf("server.address should contain port, got %s", serverAddrStr))
		}
	}, 1)
}

func testServerAddressWithDifferentPorts(ctx context.Context) {
	client1, server1 := NewMockOllamaGenerateForInvoke(ctx)
	defer server1.Close()

	client2, server2 := NewMockOllamaChatForInvoke(ctx)
	defer server2.Close()

	streamFlag := false
	req1 := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Test 1",
		Stream: &streamFlag,
	}

	req2 := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Test 2"},
		},
		Stream: &streamFlag,
	}

	err1 := client1.Generate(ctx, req1, func(resp api.GenerateResponse) error {
		return nil
	})
	if err1 != nil {
		panic(err1)
	}

	err2 := client2.Chat(ctx, req2, func(resp api.ChatResponse) error {
		return nil
	})
	if err2 != nil {
		panic(err2)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		uniqueAddrs := make(map[string]struct{})
		for _, trace := range stubs {
			for _, span := range trace {
				addr := getAttributeValue(span, "server.address")
				if addr == nil {
					continue
				}
				addrStr, ok := addr.(string)
				if !ok {
					panic("server.address should be string")
				}
				if addrStr == "" {
					panic("server.address should not be empty")
				}
				uniqueAddrs[addrStr] = struct{}{}
			}
		}

		if len(uniqueAddrs) < 2 {
			panic("Different servers should have different addresses")
		}
	}, 1)
}
