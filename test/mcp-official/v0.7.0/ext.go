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

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// GreetArgs defines the arguments for the hello_world tool.
type GreetArgs struct {
	Name string `json:"name" jsonschema:"the person to greet"`
}

// helloHandler is the handler for the hello_world tool.
func helloHandler(ctx context.Context, req *mcp.CallToolRequest, args GreetArgs) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Hello, %s!", args.Name)},
		},
	}, nil, nil
}

// findSpan searches for a span by name and kind across all traces.
func findSpan(stubs []tracetest.SpanStubs, name string, kind trace.SpanKind) tracetest.SpanStub {
	for _, t := range stubs {
		for _, span := range t {
			if span.Name == name && span.SpanKind == kind {
				return span
			}
		}
	}
	panic(fmt.Sprintf("span not found: name=%s, kind=%d", name, kind))
}
