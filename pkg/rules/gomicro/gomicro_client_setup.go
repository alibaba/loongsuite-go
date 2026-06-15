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
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	micro "go-micro.dev/v5"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/metadata"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/selector"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type clientV5Wrapper struct {
	client.Client
}

func NewV5ClientWrapper(cli client.Client) client.Client {
	return &clientV5Wrapper{cli}
}

// Call is used for client calls
func (s *clientV5Wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if !goMicroEnabler.Enable() {
		return s.Client.Call(ctx, req, rsp, opts...)
	}
	request := goMicroRequest{
		request: req,
		reqType: CallRequest,
		ctx:     ctx,
	}
	ctx = goMicroClientInstrument.Start(ctx, request)
	mda, _ := metadata.FromContext(request.ctx)
	md := metadata.Copy(mda)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(md))
	ctx = metadata.NewContext(ctx, md)
	request.ctx = ctx
	err := s.Client.Call(request.ctx, req, rsp, opts...)
	response := goMicroResponse{
		response: rsp,
		err:      err,
		ctx:      request.ctx,
	}
	goMicroClientInstrument.End(ctx, request, response, err)
	return err
}

func (s *clientV5Wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	if !goMicroEnabler.Enable() {
		return s.Client.Stream(ctx, req, opts...)
	}
	request := goMicroRequest{
		request: req,
		reqType: StreamRequest,
		ctx:     ctx,
	}
	ctx = goMicroClientInstrument.Start(ctx, request)
	stream, err := s.Client.Stream(ctx, req, opts...)
	response := goMicroResponse{
		response: stream,
		err:      err,
		ctx:      ctx,
	}
	goMicroClientInstrument.End(ctx, request, response, err)

	return stream, err

}

func (s *clientV5Wrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	return s.Client.Publish(ctx, p, opts...)

}

//go:linkname NewServiceOnEnter go-micro.dev/v5.NewServiceOnEnter
func NewServiceOnEnter(call api.CallContext, opts ...micro.Option) {
	opts = append(opts, micro.WrapClient(NewV5ClientWrapper))
	call.SetParam(0, opts)
}

//go:linkname NextOnExit go-micro.dev/v5/client.NextOnExit
func NextOnExit(call api.CallContext, nextSelector selector.Next, e error) {
	span := sdktrace.SpanFromGLS()
	if nextSelector != nil && span != nil {
		var selectWrapper selector.Next = func() (*registry.Node, error) {
			node, tmp := nextSelector()
			if node != nil {
				span.SetAttributes(semconv.ServerAddressKey.String(node.Address))
			}
			return node, tmp
		}
		call.SetReturnVal(0, selectWrapper)
	}
}
