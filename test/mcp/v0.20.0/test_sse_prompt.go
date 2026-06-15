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
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"time"
)

func main() {
	mcpServer := server.NewMCPServer("test", "1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)
	mcpServer.AddPrompt(mcp.NewPrompt(SIMPLE,
		mcp.WithPromptDescription("A simple prompt"),
	), handleSimplePrompt)

	testServer := server.NewTestServer(mcpServer)
	defer testServer.Close()
	// Connect to SSE endpoint
	c, err := client.NewSSEMCPClient(testServer.URL + "/sse")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Start the mcpClient
	if err := c.Start(ctx); err != nil {
		panic(err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		panic(err)
	}

	GetPrompt := mcp.GetPromptRequest{}
	GetPrompt.Params.Name = SIMPLE
	_, err = c.GetPrompt(ctx, GetPrompt)
	if err != nil {
		panic(err)
	}

	promptsRequest := mcp.ListPromptsRequest{}
	_, err = c.ListPrompts(ctx, promptsRequest)
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMCommonAttributes(stubs[1][0], "execute_other:initialize", "mcp", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[1][3], "execute_other:initialize", "mcp", trace.SpanKindServer)
		verifier.VerifyLLMCommonAttributes(stubs[3][0], "execute_other:prompts/get", "mcp", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[3][3], "execute_other:prompts/get", "mcp", trace.SpanKindServer)
		verifier.VerifyLLMCommonAttributes(stubs[4][0], "execute_other:prompts/list", "mcp", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[4][3], "execute_other:prompts/list", "mcp", trace.SpanKindServer)
	}, 5)
}
