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

package mcp

import (
	"context"
	"fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:linkname hookBeforeAnyOnEnter github.com/mark3labs/mcp-go/server.hookBeforeAnyOnEnter
func hookBeforeAnyOnEnter(call api.CallContext, c *server.Hooks,
	ctx context.Context, id any, method mcp.MCPMethod, message any) {
	if method == mcp.MethodPing {
		return
	}
	request := mcpRequest{
		operationName: "execute_other:" + string(method),
		system:        "mcp",
		methodType:    string(method),
		CallId:        fmt.Sprintf("%v", id),
		input:         map[string]any{},
		output:        map[string]any{},
	}
	//var subRequest *mcp.Request
	subRequest := getSubRequest(method, &request, message)
	if subRequest == nil {
		return
	}
	Ctx := ServerInstrumenter.Start(ctx, request)
	//subRequest.OtelRequest = request
	subRequest.OtelContext = Ctx
}

//go:linkname hookOnSuccessOnEnter github.com/mark3labs/mcp-go/server.hookOnSuccessOnEnter
func hookOnSuccessOnEnter(call api.CallContext, c *server.Hooks,
	ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
	subRequest := getSubRequest(method, nil, message)
	if subRequest == nil {
		return
	}
	request := mcpRequest{}
	if subRequest.OtelContext == nil {
		return
	}
	ctx, ok := subRequest.OtelContext.(context.Context)
	if !ok {
		return
	}
	ServerInstrumenter.End(ctx, request, nil, nil)
}

//go:linkname hookOnErrorOnEnter github.com/mark3labs/mcp-go/server.hookOnErrorOnEnter
func hookOnErrorOnEnter(call api.CallContext, c *server.Hooks,
	ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
	subRequest := getSubRequest(method, nil, message)
	if subRequest == nil {
		return
	}
	if subRequest.OtelContext == nil {
		return
	}
	ctx, ok := subRequest.OtelContext.(context.Context)
	if !ok {
		return
	}
	request := mcpRequest{}
	ServerInstrumenter.End(ctx, request, nil, err)
}

func getSubRequest(method mcp.MCPMethod, request *mcpRequest, message any) *mcp.Request {
	switch method {
	case mcp.MethodToolsCall:
		if msg, ok := message.(*mcp.CallToolRequest); ok {
			if request != nil {
				request.operationName = "execute_tool"
				request.methodName = msg.Params.Name
			}
			return &msg.Request
		}
	case mcp.MethodPromptsGet:
		if msg, ok := message.(*mcp.GetPromptRequest); ok {
			if request != nil {
				request.input["prompt_name"] = msg.Params.Name
			}
			return &msg.Request
		}
	case mcp.MethodResourcesRead:
		if msg, ok := message.(*mcp.ReadResourceRequest); ok {
			if request != nil {
				request.input["resources_uri"] = msg.Params.URI
			}
			return &msg.Request
		}
	case mcp.MethodInitialize:
		if msg, ok := message.(*mcp.InitializeRequest); ok {
			if request != nil {
				request.input["client_info_name"] = msg.Params.ClientInfo.Name
				request.input["client_info_version"] = msg.Params.ClientInfo.Version
			}
			return &msg.Request
		}
	case mcp.MethodResourcesList:
		if msg, ok := message.(*mcp.ListResourcesRequest); ok {
			if request != nil {
				request.input["cursor"] = msg.Params.Cursor
			}
			return &msg.Request
		}
	case mcp.MethodResourcesTemplatesList:
		if msg, ok := message.(*mcp.ListResourceTemplatesRequest); ok {
			if request != nil {
				request.input["cursor"] = msg.Params.Cursor
			}
			return &msg.Request
		}
	case mcp.MethodPromptsList:
		if msg, ok := message.(*mcp.ListPromptsRequest); ok {
			if request != nil {
				request.input["cursor"] = msg.Params.Cursor
			}
			return &msg.Request
		}
	case mcp.MethodToolsList:
		if msg, ok := message.(*mcp.ListToolsRequest); ok {
			if request != nil {
				request.input["cursor"] = msg.Params.Cursor
			}
			return &msg.Request
		}
	}
	return nil
}
