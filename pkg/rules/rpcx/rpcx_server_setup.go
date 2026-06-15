// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rpcx

import (
	"context"
	_ "unsafe"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

var rpcxServerInstrumenter = BuildRpcxServerInstrumenter()

// func (s *Server) handleRequest(ctx context.Context, req *protocol.Message) (res *protocol.Message, err error)
//
//go:linkname serverHandleRequestOnEnter github.com/smallnest/rpcx/server.serverHandleRequestOnEnter
func serverHandleRequestOnEnter(call api.CallContext, s *server.Server, ctx context.Context, req *protocol.Message) {
	if !rpcxEnabler.Enable() {
		return
	}

	xcall := new(client.Call)
	xcall.ServicePath = req.ServicePath
	xcall.ServiceMethod = req.ServiceMethod
	xcall.Metadata = req.Metadata

	request := rpcxReq{
		call: xcall,
		addr: s.Address().String(),
		ctx:  ctx,
		withCtx: func(_ctx context.Context) {
			call.SetParam(1, _ctx)
		},
	}
	newCtx := rpcxServerInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = newCtx
	data["request"] = request
	call.SetData(data)
}

// func (s *Server) handleRequest(ctx context.Context, req *protocol.Message) (res *protocol.Message, err error)
//
//go:linkname serverHandleRequestOnExit github.com/smallnest/rpcx/server.serverHandleRequestOnExit
func serverHandleRequestOnExit(call api.CallContext, res *protocol.Message, err error) {
	if !rpcxEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)
	request := data["request"].(rpcxReq)
	rpcxServerInstrumenter.End(ctx, request, rpcxRes{}, err)
}
