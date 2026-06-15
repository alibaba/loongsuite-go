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
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

// Since standard input/output communication cannot be implemented in this test, it will hang the test process. Therefore, this method is only reserved for future use when stdio is available and is not used currently.
// Since standard input/output communication cannot be implemented in the test here and will cause the test process to hang, this method is reserved for future use when stdio becomes available, and is currently not used.

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := client.NewStdioMCPClient(
		"./test_stdio_server",
		[]string{"IN_OTEL_TEST=true"}, // Empty ENV
	)
	if err != nil {
		panic(err)
	}
	defer c.Close()
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
	listTmpRequest := mcp.CallToolRequest{}
	listTmpRequest.Params.Name = "hello_world"
	listTmpRequest.Params.Arguments = map[string]interface{}{
		"name": "abc",
	}
	_, err = c.CallTool(ctx, listTmpRequest)
	if err != nil {
		panic(err)
	}
	toolsRequest := mcp.ListToolsRequest{}
	_, err = c.ListTools(ctx, toolsRequest)
	if err != nil {
		panic(err)
	}
	/*verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMCommonAttributes(stubs[0][2], "execute_other:initialize", "mcp")
		verifier.VerifyLLMCommonAttributes(stubs[3][2], "execute_tool", "mcp")
		verifier.VerifyLLMCommonAttributes(stubs[4][2], "execute_other:tools/list", "mcp")
	}, 0)*/
}
