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
	"fmt"
	"os"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/rpc"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
)

type rpcxInnerEnabler struct {
	enabled bool
}

func (t rpcxInnerEnabler) Enable() bool {
	return t.enabled
}

var rpcxEnabler = rpcxInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_RPCX_ENABLED") != "false"}

type rpcxClientAttrsGetter struct {
}

func (t rpcxClientAttrsGetter) GetSystem(request rpcxReq) string {
	return "rpcx"
}

func (t rpcxClientAttrsGetter) GetService(request rpcxReq) string {
	return request.call.ServicePath
}

func (t rpcxClientAttrsGetter) GetMethod(request rpcxReq) string {
	return request.call.ServiceMethod
}

func (t rpcxClientAttrsGetter) GetServerAddress(request rpcxReq) string {
	return request.addr
}

type rpcxServerAttrsGetter struct {
}

func (t rpcxServerAttrsGetter) GetSystem(request rpcxReq) string {
	return "rpcx"
}

func (t rpcxServerAttrsGetter) GetService(request rpcxReq) string {
	return request.call.ServicePath
}

func (t rpcxServerAttrsGetter) GetMethod(request rpcxReq) string {
	return request.call.ServiceMethod
}

func (t rpcxServerAttrsGetter) GetServerAddress(request rpcxReq) string {
	return request.addr
}

type rpcxStatusCodeExtractor[REQUEST rpcxReq, RESPONSE rpcxRes] struct {
}

func (t rpcxStatusCodeExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request rpcxReq, response rpcxRes, err error) {
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("rpcx request error.%v", err))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type rpcxRequestCarrier struct {
	call    *client.Call
	ctx     context.Context
	withCtx func(ctx context.Context)
}

func (t rpcxRequestCarrier) Get(key string) string {
	return t.call.Metadata[key]
}

func (t rpcxRequestCarrier) Set(key string, value string) {
	md := t.ctx.Value(share.ReqMetaDataKey)
	if md == nil {
		md = map[string]string{}
	}
	if _, ok := md.(map[string]string)[key]; ok {
		return
	}
	md.(map[string]string)[key] = value
	if t.withCtx != nil {
		t.withCtx(context.WithValue(t.ctx, share.ReqMetaDataKey, md))
	}
}

func (t rpcxRequestCarrier) Keys() []string {
	var vals = make([]string, 0)
	md := t.ctx.Value(share.ReqMetaDataKey)
	if md != nil {
		for k := range md.(map[string]string) {
			vals = append(vals, k)
		}
	}
	return vals
}

func BuildRpcxClientInstrumenter() instrumenter.Instrumenter[rpcxReq, rpcxRes] {
	builder := instrumenter.Builder[rpcxReq, rpcxRes]{}
	clientGetter := rpcxClientAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&rpcxStatusCodeExtractor[rpcxReq, rpcxRes]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[rpcxReq]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[rpcxReq]{}).
		AddAttributesExtractor(&rpc.ClientRpcAttrsExtractor[rpcxReq, rpcxRes, rpcxClientAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.RPCXGO_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcClientMetrics("rpcx.client")).
		BuildPropagatingToDownstreamInstrumenter(
			func(n rpcxReq) propagation.TextMapCarrier {
				return rpcxRequestCarrier{ctx: n.ctx, call: n.call, withCtx: n.withCtx}
			},
			otel.GetTextMapPropagator(),
		)
}

func BuildRpcxServerInstrumenter() instrumenter.Instrumenter[rpcxReq, rpcxRes] {
	builder := instrumenter.Builder[rpcxReq, rpcxRes]{}
	serverGetter := rpcxServerAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&rpcxStatusCodeExtractor[rpcxReq, rpcxRes]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[rpcxReq]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[rpcxReq]{}).
		AddAttributesExtractor(&rpc.ServerRpcAttrsExtractor[rpcxReq, rpcxRes, rpcxServerAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.RPCXGO_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcServerMetrics("rpcx.server")).
		BuildPropagatingFromUpstreamInstrumenter(
			func(n rpcxReq) propagation.TextMapCarrier {
				return rpcxRequestCarrier{ctx: n.ctx, call: n.call, withCtx: n.withCtx}
			},
			otel.GetTextMapPropagator(),
		)
}
