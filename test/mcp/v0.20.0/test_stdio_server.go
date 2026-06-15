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
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Since standard input/output communication cannot be implemented in this test, it will hang the test process. Therefore, this method is only reserved for future use when stdio is available and is not used currently.
// Since standard input/output communication cannot be implemented in the test here and will cause the test process to hang, this method is reserved for future use when stdio becomes available, and is currently not used.

func main() {
	mcpServer := server.NewMCPServer("test", "1.0.0",
		//server.WithResourceCapabilities(true, true),
		//server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)
	tool := mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)
	// Add tool handler
	mcpServer.AddTool(tool, helloHandler)

	mcpServer.AddPrompt(mcp.NewPrompt(SIMPLE,
		mcp.WithPromptDescription("A simple prompt"),
	), handleSimplePrompt)
	if err := server.ServeStdio(mcpServer); err != nil {
		panic(err)
	}

	mcpServer.AddResource(mcp.NewResource("test://static/resource",
		"Static Resource",
		mcp.WithMIMEType("text/plain"),
	), handleReadResource)

	mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"test://dynamic/resource/{id}",
			"Dynamic Resource",
		),
		handleResourceTemplate,
	)
	if err := server.ServeStdio(mcpServer); err != nil {
		panic(err)
	}
}
