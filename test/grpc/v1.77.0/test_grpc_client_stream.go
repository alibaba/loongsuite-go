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

package main

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"time"
)

func main() {
	// starter server
	go setupGRPC()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	sendStreamReq(context.Background())
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		// filter out client
		verifier.Assert(len(stubs[0]) == 1, "Except client grpc system to be 1, got %d", len(stubs))
		verifier.Assert(stubs[0][0].SpanKind == trace.SpanKindServer, "Except client grpc system to be server, got %v", stubs[0][0].SpanKind)
	}, 1)
}
