// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package grpc

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

//go:linkname grpcClientOnEnter google.golang.org/grpc.grpcClientOnEnter
func grpcClientOnEnter(call api.CallContext, ctx context.Context, target string, opts ...grpc.DialOption) {
	if !grpcEnabler.Enable() {
		return
	}
	h := grpc.WithStatsHandler(NewClientHandler())
	var opt []grpc.DialOption
	opt = append(opt, h)
	opt = append(opt, opts...)
	call.SetParam(2, opt)
}

//go:linkname grpcClientOnExit google.golang.org/grpc.grpcClientOnExit
func grpcClientOnExit(call api.CallContext, cc *grpc.ClientConn, err error) {
	if !grpcEnabler.Enable() {
		return
	}
	return
}

type clientHandler struct {
	serverAddr string
	*grpcOtelConfig
}

func NewClientHandler(opts ...Option) stats.Handler {
	h := &clientHandler{
		grpcOtelConfig: newConfig(opts, "client"),
	}

	return h
}

// TagRPC can attach some information to the given context.
func (h *clientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if info.FullMethodName == grpcTraceExporterPath || info.FullMethodName == grpcMetricExporterPath {
		return ctx
	}
	nCtx := grpcClientInstrument.Start(ctx, grpcRequest{
		methodName:    info.FullMethodName,
		serverAddress: h.serverAddr,
	})
	gCtx := gRPCContext{
		methodName: info.FullMethodName,
	}

	return inject(context.WithValue(nCtx, gRPCContextKey{}, &gCtx), h.grpcOtelConfig.Propagators, info.FullMethodName)
}

// HandleRPC processes the RPC stats.
func (h *clientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	h.handleRPC(ctx, rs, false)
}

// TagConn can attach some information to the given context.
func (h *clientHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	h.serverAddr = info.RemoteAddr.String()
	return ctx
}

// HandleConn processes the Conn stats.
func (h *clientHandler) HandleConn(context.Context, stats.ConnStats) {
	// no-op
}
