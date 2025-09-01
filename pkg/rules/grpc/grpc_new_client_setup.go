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

//go:linkname grpcNewClientOnEnter google.golang.org/grpc.grpcNewClientOnEnter
func grpcNewClientOnEnter(call api.CallContext, target string, opts ...grpc.DialOption) {
	h := grpc.WithStatsHandler(NewClientNewHandler())
	var opt []grpc.DialOption
	opt = append(opt, h)
	opt = append(opt, opts...)
	call.SetParam(1, opt)
}

//go:linkname grpcNewClientOnExit google.golang.org/grpc.grpcNewClientOnExit
func grpcNewClientOnExit(call api.CallContext, cc *grpc.ClientConn, err error) {
	return
}

type clientNewHandler struct {
	serverAddr string
	*grpcOtelConfig
}

func NewClientNewHandler(opts ...Option) stats.Handler {
	h := &clientNewHandler{
		grpcOtelConfig: newConfig(opts, "client"),
	}

	return h
}

// TagRPC can attach some information to the given context.
func (h *clientNewHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	if info.FullMethodName == grpcTraceExporterPath || info.FullMethodName == grpcMetricExporterPath {
		return ctx
	}
	filter, _ := ctx.Value("stream_filter").(*bool)
	if filter != nil && *filter {
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
func (h *clientNewHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	h.handleRPC(ctx, rs, false)
}

// TagConn can attach some information to the given context.
func (h *clientNewHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	h.serverAddr = info.RemoteAddr.String()
	return ctx
}

// HandleConn processes the Conn stats.
func (h *clientNewHandler) HandleConn(context.Context, stats.ConnStats) {
	// no-op
}
