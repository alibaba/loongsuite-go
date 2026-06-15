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

package mcp0_20_0

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

//go:linkname clientSseOnEnter github.com/mark3labs/mcp-go/client.clientSseOnEnter
func clientSseOnEnter(call api.CallContext, c *client.SSEMCPClient,
	ctx context.Context,
	method string,
	params interface{}) {
	clientOnEnter(call, ctx, method, params)
}

//go:linkname clientStdioOnEnter github.com/mark3labs/mcp-go/client.clientStdioOnEnter
func clientStdioOnEnter(call api.CallContext, c *client.StdioMCPClient,
	ctx context.Context,
	method string,
	params interface{}) {
	clientOnEnter(call, ctx, method, params)
}

func clientOnEnter(call api.CallContext,
	ctx context.Context,
	method string,
	params interface{}) {
	if method == string(mcp.MethodPing) {
		return
	}
	request := mcpRequest{
		operationName: "execute_other:" + string(method),
		system:        "mcp",
		methodType:    method,
		input:         map[string]any{},
		output:        map[string]any{},
	}
	//var subRequest *mcp.Request
	if err := handleClientRequest(method, &request, params); err != nil {
		fmt.Println("handleClientRequest", "未匹配")
		return
	}
	Ctx := ClientInstrumenter.Start(ctx, request)
	data := make(map[string]interface{})
	data["ctx"] = Ctx
	data["mcp_client_request"] = request
	call.SetData(data)
}

//go:linkname clientSseOnExit github.com/mark3labs/mcp-go/client.clientSseOnExit
func clientSseOnExit(call api.CallContext, j *json.RawMessage, err error) {
	clientOnExit(call, j, err)
}

//go:linkname clientStdioOnExit github.com/mark3labs/mcp-go/client.clientStdioOnExit
func clientStdioOnExit(call api.CallContext, j *json.RawMessage, err error) {
	clientOnExit(call, j, err)
}
func clientOnExit(call api.CallContext, j *json.RawMessage, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	request, ok := data["mcp_client_request"].(mcpRequest)
	if !ok {
		return
	}
	ClientInstrumenter.End(ctx, request, nil, err)
}

func handleClientRequest(method string, request *mcpRequest, message interface{}) error {
	switch method {
	case string(mcp.MethodToolsCall):
		if msg, ok := message.(struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}); ok {
			if request != nil {
				request.operationName = "execute_tool"
				request.methodName = msg.Name
			}
		}
		return nil
	case string(mcp.MethodPromptsGet):
		if msg, ok := message.(struct {
			// The name of the prompt or prompt template.
			Name string `json:"name"`
			// Arguments to use for templating the prompt.
			Arguments map[string]string `json:"arguments,omitempty"`
		}); ok {
			if request != nil {
				request.input["prompt_name"] = msg.Name
			}
		}
		return nil
	case string(mcp.MethodResourcesRead):
		if msg, ok := message.(struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}); ok {
			if request != nil {
				request.input["resources_uri"] = msg.URI
			}
		}
		return nil
	case string(mcp.MethodInitialize):
		if msg, ok := message.(struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    mcp.ClientCapabilities `json:"capabilities"`
			ClientInfo      mcp.Implementation     `json:"clientInfo"`
		}); ok {
			if request != nil {
				request.input["client_info_name"] = msg.ClientInfo.Name
				request.input["client_info_version"] = msg.ClientInfo.Version
			}
		}
		return nil
	case string(mcp.MethodResourcesList):
		if msg, ok := message.(struct {
			Cursor mcp.Cursor `json:"cursor,omitempty"`
		}); ok {
			if request != nil {
				request.input["cursor"] = msg.Cursor
			}
		}
		return nil
	case string(mcp.MethodResourcesTemplatesList):
		if msg, ok := message.(struct {
			Cursor mcp.Cursor `json:"cursor,omitempty"`
		}); ok {
			if request != nil {
				request.input["cursor"] = msg.Cursor
			}
		}
		return nil
	case string(mcp.MethodPromptsList):
		if msg, ok := message.(struct {
			Cursor mcp.Cursor `json:"cursor,omitempty"`
		}); ok {
			if request != nil {
				request.input["cursor"] = msg.Cursor
			}
		}
		return nil
	case string(mcp.MethodToolsList):
		if msg, ok := message.(struct {
			Cursor mcp.Cursor `json:"cursor,omitempty"`
		}); ok {
			if request != nil {
				request.input["cursor"] = msg.Cursor
			}
		}
		return nil
	}
	return errors.New("client method not match")
}
