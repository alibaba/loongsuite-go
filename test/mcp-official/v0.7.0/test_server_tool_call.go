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

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()

	// Create server with a tool
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello_world",
		Description: "Say hello to someone",
	}, helloHandler)

	// Create client
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)

	// Connect using in-memory transport
	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, t1, nil)
	if err != nil {
		panic(err)
	}
	defer ss.Close()

	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Call tool
	_, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "hello_world",
		Arguments: map[string]any{"name": "world"},
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		xx, _ := json.Marshal(stubs)
		fmt.Println(string(xx))

		// Verify server initialize span
		serverInitSpan := findSpan(stubs, "execute_other:initialize", trace.SpanKindServer)
		verifier.VerifyLLMCommonAttributes(serverInitSpan, "execute_other:initialize", "mcp", trace.SpanKindServer)

		// Verify server tool call span
		serverToolSpan := findSpan(stubs, "execute_tool", trace.SpanKindServer)
		verifier.VerifyLLMCommonAttributes(serverToolSpan, "execute_tool", "mcp", trace.SpanKindServer)
	}, 4)
}
