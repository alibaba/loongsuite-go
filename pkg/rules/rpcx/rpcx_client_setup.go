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

package rpcx

import (
	"context"
	"log"
	_ "unsafe"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

var (
	rpcxClientInstrumenter = BuildRpcxClientInstrumenter()
	iscallFuncKey          = "__is_call_func__"
)

// func (client *Client) call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
//
//go:linkname clientRpcxCallOnEnter github.com/smallnest/rpcx/client.clientRpcxCallOnEnter
func clientRpcxCallOnEnter(call api.CallContext, cli *client.Client, ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) {
	if !rpcxEnabler.Enable() {
		return
	}
	if cli == nil {
		return
	}
	xcall := new(client.Call)
	xcall.ServicePath = servicePath
	xcall.ServiceMethod = serviceMethod

	ctx = context.WithValue(ctx, iscallFuncKey, true)
	req := rpcxReq{
		call: xcall,
		addr: cli.RemoteAddr(),
		ctx:  ctx,
		withCtx: func(_ctx context.Context) {
			call.SetParam(1, _ctx)
		},
	}
	newCtx := rpcxClientInstrumenter.Start(ctx, req)
	data := make(map[string]interface{}, 2)
	data["ctx"] = newCtx
	data["request"] = req
	call.SetData(data)

	logPrintf("Request", servicePath, serviceMethod, nil)
}

// func (client *Client) call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
//
//go:linkname clientRpcxCallOnExit github.com/smallnest/rpcx/client.clientRpcxCallOnExit
func clientRpcxCallOnExit(call api.CallContext, err error) {
	if !rpcxEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)
	request := data["request"].(rpcxReq)
	logPrintf("Response", request.call.ServicePath, request.call.ServiceMethod, err)

	rpcxClientInstrumenter.End(ctx, request, rpcxRes{}, err)

}

// func (client *Client) Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call
//
//go:linkname clientRpcxGoOnEnter github.com/smallnest/rpcx/client.clientRpcxGoOnEnter
func clientRpcxGoOnEnter(call api.CallContext, cli *client.Client, ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *client.Call) {
	if !rpcxEnabler.Enable() {
		return
	}
	if cli == nil {
		return
	}
	// The call method will invoke the Go method, which will cause the span to be duplicated. Therefore, there is no need to create the span again
	if v := ctx.Value(iscallFuncKey); v != nil && v.(bool) {
		data := make(map[string]interface{}, 1)
		data["ctx"] = ctx
		call.SetData(data)
		return
	}
	xcall := new(client.Call)
	xcall.ServicePath = servicePath
	xcall.ServiceMethod = serviceMethod

	req := rpcxReq{
		call: xcall,
		addr: cli.RemoteAddr(),
		ctx:  ctx,
		withCtx: func(_ctx context.Context) {
			call.SetParam(1, _ctx)
		},
	}
	newCtx := rpcxClientInstrumenter.Start(ctx, req)
	data := make(map[string]interface{}, 2)
	data["ctx"] = newCtx
	data["request"] = req
	call.SetData(data)

	logPrintf("Request", servicePath, serviceMethod, nil)
}

// func (client *Client) Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call
//
//go:linkname clientRpcxGoOnExit github.com/smallnest/rpcx/client.clientRpcxGoOnExit
func clientRpcxGoOnExit(call api.CallContext, xcall *client.Call) {
	if !rpcxEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)

	if v := ctx.Value(iscallFuncKey); v != nil && v.(bool) {
		return
	}
	request := data["request"].(rpcxReq)
	var err error
	if xcall != nil {
		err = xcall.Error
	}
	logPrintf("Response", request.call.ServicePath, request.call.ServiceMethod, err)

	rpcxClientInstrumenter.End(ctx, request, rpcxRes{}, err)
}

// (client *Client) SendRaw(ctx context.Context, r *protocol.Message) (map[string]string, []byte, error)
//
//go:linkname clientRpcxSendRawOnEnter github.com/smallnest/rpcx/client.clientRpcxSendRawOnEnter
func clientRpcxSendRawOnEnter(call api.CallContext, cli *client.Client, ctx context.Context, r *protocol.Message) {
	if !rpcxEnabler.Enable() {
		return
	}
	if cli == nil {
		return
	}
	xcall := new(client.Call)
	xcall.ServicePath = r.ServicePath
	xcall.ServiceMethod = r.ServiceMethod

	req := rpcxReq{
		call: xcall,
		addr: cli.RemoteAddr(),
		ctx:  ctx,
		withCtx: func(_ctx context.Context) {
			call.SetParam(1, _ctx)
		},
	}
	newCtx := rpcxClientInstrumenter.Start(ctx, req)
	data := make(map[string]interface{}, 2)
	data["ctx"] = newCtx
	data["request"] = req
	call.SetData(data)

	logPrintf("Request", r.ServicePath, r.ServiceMethod, nil)
}

// (client *Client) SendRaw(ctx context.Context, r *protocol.Message) (map[string]string, []byte, error)
//
//go:linkname clientRpcxSendRawOnExit github.com/smallnest/rpcx/client.clientRpcxSendRawOnExit
func clientRpcxSendRawOnExit(call api.CallContext, _ map[string]string, _ []byte, err error) {
	if !rpcxEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)
	request := data["request"].(rpcxReq)
	logPrintf("Response", request.call.ServicePath, request.call.ServiceMethod, err)
	rpcxClientInstrumenter.End(ctx, request, rpcxRes{}, err)
}

func logPrintf(state, servicePath, serviceMethod string, err error) {
	if err != nil {
		log.Printf(` state=%s rpc.service=%s rpc.method=%s err=%v`, state, servicePath, serviceMethod, err)
	} else {
		log.Printf(` state=%s rpc.service=%s rpc.method=%s`, state, servicePath, serviceMethod)
	}
}
