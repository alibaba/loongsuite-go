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

package dubbo

import (
	"context"
	_ "unsafe"

	"dubbo.apache.org/dubbo-go/v3/protocol"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel"
)

var dubboClientInstrumenter = BuildDubboClientInstrumenter()

//go:linkname dubboConsumerGracefulShutdownFilterInvokeOnEnter dubbo.apache.org/dubbo-go/v3/filter/graceful_shutdown.dubboConsumerGracefulShutdownFilterInvokeOnEnter
func dubboConsumerGracefulShutdownFilterInvokeOnEnter(call api.CallContext, _ interface{}, ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) {
	if !dubboEnabler.Enable() {
		return
	}

	req := dubboRequest{
		methodName:    invocation.MethodName(),
		serviceKey:    invoker.GetURL().ServiceKey(),
		serverAddress: invoker.GetURL().Address(),
	}

	ctx = dubboClientInstrumenter.Start(ctx, req)

	attachments := invocation.Attachments()
	if attachments == nil {
		attachments = map[string]any{}
	}
	inject(ctx, attachments, otel.GetTextMapPropagator())
	for k, v := range attachments {
		invocation.SetAttachment(k, v)
	}

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["req"] = req

	call.SetData(data)
}

//go:linkname dubboConsumerGracefulShutdownFilterInvokeOnExit dubbo.apache.org/dubbo-go/v3/filter/graceful_shutdown.dubboConsumerGracefulShutdownFilterInvokeOnExit
func dubboConsumerGracefulShutdownFilterInvokeOnExit(call api.CallContext, res protocol.Result) {
	if !dubboEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	req, ok := data["req"].(dubboRequest)
	if !ok {
		return
	}

	resp := dubboResponse{}
	if res.Error() != nil {
		resp.hasError = true
		resp.errorMsg = res.Error().Error()
	}

	dubboClientInstrumenter.End(ctx, req, resp, res.Error())
}
