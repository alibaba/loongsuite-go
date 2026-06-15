// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package kitex

import (
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"os"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/rpc"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
)

type kitexInnerEnabler struct {
	enabled bool
}

func (k kitexInnerEnabler) Enable() bool {
	return k.enabled
}

var kitexEnabler = kitexInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_KITEX_ENABLED") != "false"}

type kitexAttrsGetter struct{}

func (g kitexAttrsGetter) GetSystem(request rpcinfo.RPCInfo) string {
	return "kitex"
}

func (g kitexAttrsGetter) GetService(ri rpcinfo.RPCInfo) string {
	if ri.Invocation().PackageName() != "" {
		return ri.Invocation().PackageName() + "." + ri.Invocation().ServiceName()
	}
	return ri.Invocation().ServiceName()
}

func (g kitexAttrsGetter) GetMethod(ri rpcinfo.RPCInfo) string {
	if ri.Invocation().PackageName() != "" {
		return ri.Invocation().PackageName() + "." + ri.Invocation().ServiceName() + "/" + ri.Invocation().MethodName()
	}
	return ri.Invocation().ServiceName() + "/" + ri.Invocation().MethodName()
}

func (g kitexAttrsGetter) GetServerAddress(request rpcinfo.RPCInfo) string {
	if request.To() != nil && request.To().Address() != nil {
		return request.To().Address().String()
	}
	return ""
}

func BuildKitexClientInstrumenter() instrumenter.Instrumenter[rpcinfo.RPCInfo, rpcinfo.RPCInfo] {
	builder := instrumenter.Builder[rpcinfo.RPCInfo, rpcinfo.RPCInfo]{}
	clientGetter := kitexAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[rpcinfo.RPCInfo]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[rpcinfo.RPCInfo]{}).
		AddAttributesExtractor(&rpc.ClientRpcAttrsExtractor[rpcinfo.RPCInfo, rpcinfo.RPCInfo, kitexAttrsGetter]{Base: rpc.RpcAttrsExtractor[rpcinfo.RPCInfo, rpcinfo.RPCInfo, kitexAttrsGetter]{Getter: clientGetter}}).
		AddOperationListeners(rpc.RpcClientMetrics("kitex.client")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.KITEX_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

func BuildKitexServerInstrumenter() instrumenter.Instrumenter[rpcinfo.RPCInfo, rpcinfo.RPCInfo] {
	builder := instrumenter.Builder[rpcinfo.RPCInfo, rpcinfo.RPCInfo]{}
	serverGetter := kitexAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[rpcinfo.RPCInfo]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[rpcinfo.RPCInfo]{}).
		AddOperationListeners(rpc.RpcServerMetrics("kitex.server")).
		AddAttributesExtractor(&rpc.ServerRpcAttrsExtractor[rpcinfo.RPCInfo, rpcinfo.RPCInfo, kitexAttrsGetter]{Base: rpc.RpcAttrsExtractor[rpcinfo.RPCInfo, rpcinfo.RPCInfo, kitexAttrsGetter]{Getter: serverGetter}}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.KITEX_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

func parseRPCError(ri rpcinfo.RPCInfo) (panicMsg, panicStack string, err error) {
	panicked, panicErr := ri.Stats().Panicked()
	if err = ri.Stats().Error(); err == nil && !panicked {
		return
	}
	if panicked {
		panicMsg = fmt.Sprintf("%v", panicErr)
		if stackErr, ok := panicErr.(interface{ Stack() string }); ok {
			panicStack = stackErr.Stack()
		}
	}
	return
}
