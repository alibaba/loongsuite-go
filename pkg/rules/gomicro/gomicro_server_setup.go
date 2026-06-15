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

package gomicro

import (
	"context"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go-micro.dev/v5/metadata"
	"go-micro.dev/v5/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var goMicroServerInstrument = BuildGoMicroServerOtelInstrumenter()

//go:linkname ServeRequestOnEnter go-micro.dev/v5/server.ServeRequestOnEnter
func ServeRequestOnEnter(call api.CallContext, _ interface{}, ctx context.Context, request server.Request, response server.Response) {
	if !goMicroEnabler.Enable() {
		return
	}
	propagators := otel.GetTextMapPropagator()
	mda, _ := metadata.FromContext(ctx)
	for key, val := range mda {
		mda[strings.ToLower(key)] = val
	}
	ctx = propagators.Extract(ctx, propagation.MapCarrier(mda))
	req := goMicroServerRequest{
		request: request,
		ctx:     ctx,
	}
	ctx = goMicroServerInstrument.Start(ctx, req)
	call.SetParam(1, ctx)
	data := make(map[string]interface{}, 2)
	data["ctx"] = ctx
	data["request"] = req
	call.SetData(data)
	return
}

//go:linkname ServeRequestOnExit go-micro.dev/v5/server.ServeRequestOnExit
func ServeRequestOnExit(call api.CallContext, r error) {
	if !goMicroEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request, ok := data["request"].(goMicroServerRequest)
	if !ok {
		return
	}
	response := goMicroResponse{
		response: nil,
		err:      r,
		ctx:      ctx,
	}
	goMicroServerInstrument.End(ctx, request, response, r)
}
