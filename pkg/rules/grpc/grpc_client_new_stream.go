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
)

//go:linkname grpcClientNewStreamOnEnter google.golang.org/grpc.grpcClientNewStreamOnEnter
func grpcClientNewStreamOnEnter(call api.CallContext, cc *grpc.ClientConn, ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) {
	var stream_filter bool
	stream_filter = true
	ctx = context.WithValue(ctx, "stream_filter", &stream_filter)
	call.SetParam(1, ctx)
}

//go:linkname grpcClientNewStreamOnExit google.golang.org/grpc.grpcClientNewStreamOnExit
func grpcClientNewStreamOnExit(call api.CallContext, cs grpc.ClientStream, err error) {
	return
}
