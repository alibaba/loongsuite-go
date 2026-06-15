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
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"os"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/rpc"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type grpcInnerEnabler struct {
	enabled bool
}

func (g grpcInnerEnabler) Enable() bool {
	return g.enabled
}

var grpcEnabler = grpcInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GRPC_ENABLED") != "false"}

type grpcAttrsGetter struct {
}

func (g grpcAttrsGetter) GetSystem(request grpcRequest) string {
	return "grpc"
}

func (g grpcAttrsGetter) GetService(request grpcRequest) string {
	fullMethodName := request.methodName
	slashIndex := strings.LastIndex(fullMethodName, "/")
	if slashIndex == -1 {
		return ""
	}
	return fullMethodName[0:slashIndex]
}

func (g grpcAttrsGetter) GetMethod(request grpcRequest) string {
	fullMethodName := request.methodName
	slashIndex := strings.LastIndex(fullMethodName, "/")
	if slashIndex == -1 {
		return ""
	}
	return fullMethodName[slashIndex+1:]
}

func (g grpcAttrsGetter) GetServerAddress(request grpcRequest) string {
	return request.serverAddress
}

type grpcStatusCodeExtractor[REQUEST grpcRequest, RESPONSE grpcResponse] struct {
}

func (g grpcStatusCodeExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request grpcRequest, response grpcResponse, err error) {
	statusCode := response.statusCode
	if statusCode != 200 {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Error, fmt.Sprintf("wrong grpc status code %d", statusCode))
		}
	}
}

func BuildGrpcClientInstrumenter() instrumenter.Instrumenter[grpcRequest, grpcResponse] {
	builder := instrumenter.Builder[grpcRequest, grpcResponse]{}
	clientGetter := grpcAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&grpcStatusCodeExtractor[grpcRequest, grpcResponse]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[grpcRequest]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[grpcRequest]{}).
		AddAttributesExtractor(&rpc.ClientRpcAttrsExtractor[grpcRequest, grpcResponse, grpcAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GRPC_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcClientMetrics("grpc.client")).
		BuildInstrumenter()
}

func BuildGrpcServerInstrumenter() instrumenter.Instrumenter[grpcRequest, grpcResponse] {
	builder := instrumenter.Builder[grpcRequest, grpcResponse]{}
	serverGetter := grpcAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&grpcStatusCodeExtractor[grpcRequest, grpcResponse]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[grpcRequest]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[grpcRequest]{}).
		AddAttributesExtractor(&rpc.ServerRpcAttrsExtractor[grpcRequest, grpcResponse, grpcAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GRPC_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcServerMetrics("grpc.server")).
		BuildInstrumenter()
}
