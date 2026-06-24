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

package mcpofficial

import (
	"context"
	"encoding/json"
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:linkname afterNewClient github.com/modelcontextprotocol/go-sdk/mcp.afterNewClient
func afterNewClient(call api.CallContext, c *mcp.Client) {
	if !mcpEnabler.Enable() || c == nil {
		return
	}

	monitoringMiddleware := createClientMonitoringMiddleware()
	c.AddSendingMiddleware(monitoringMiddleware)
	call.SetReturnVal(0, c)
}

// createClientMonitoringMiddleware creates a client monitoring middleware
func createClientMonitoringMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
			if !mcpEnabler.Enable() {
				return next(ctx, method, req)
			}

			request := mcpRequest{
				operationName: "execute_other:" + method,
				system:        "mcp",
				methodType:    method,
				input:         map[string]any{},
				output:        map[string]any{},
			}

			// Extract request attributes based on method type
			extractClientRequestAttributes(req, method, &request)

			// Start span via instrumenter
			ctx = ClientInstrumenter.Start(ctx, request)

			// Execute original handler
			result, err = next(ctx, method, req)

			// End span
			ClientInstrumenter.End(ctx, request, nil, err)
			return result, err
		}
	}
}

// extractClientRequestAttributes extracts client request attributes
func extractClientRequestAttributes(req mcp.Request, method string, request *mcpRequest) {
	params := req.GetParams()
	if isNilInterface(params) {
		return
	}

	// Serialize parameters
	paramsJSON, err := json.Marshal(params)
	if err == nil {
		request.input["mcp.arguments"] = string(paramsJSON)
	}

	// Extract method-specific attributes
	switch method {
	case "initialize":
		if initParams, ok := params.(*mcp.InitializeParams); ok && initParams != nil {
			if initParams.ProtocolVersion != "" {
				request.input["network.protocol.version"] = initParams.ProtocolVersion
			}
			if initParams.ClientInfo != nil {
				request.input["client_info_name"] = initParams.ClientInfo.Name
				request.input["client_info_version"] = initParams.ClientInfo.Version
			}
		}
	case "tools/call":
		if callParams, ok := params.(*mcp.CallToolParams); ok && callParams != nil {
			if callParams.Name != "" {
				request.operationName = "execute_tool"
				request.methodName = callParams.Name
			}
		}
	case "resources/read":
		if readParams, ok := params.(*mcp.ReadResourceParams); ok && readParams != nil {
			if readParams.URI != "" {
				request.input["mcp.resource.uri"] = readParams.URI
			}
		}
	case "prompts/get":
		if getParams, ok := params.(*mcp.GetPromptParams); ok && getParams != nil {
			if getParams.Name != "" {
				request.input["prompt_name"] = getParams.Name
			}
		}
	}
}

func isNilInterface(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}
