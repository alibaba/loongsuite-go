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

package main

import (
	"context"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	// starter server
	go setupDubbo()
	time.Sleep(3 * time.Second)
	// use a client to request to the server
	sendErrDubboReq(context.Background())
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyRpcClientAttributes(stubs[0][0], "greet.GreetService/Greet", "apache_dubbo", "greet.GreetService", "Greet")
		verifier.VerifyRpcServerAttributes(stubs[0][1], "greet.GreetService/Greet", "apache_dubbo", "greet.GreetService", "Greet")
		if stubs[0][0].Status.Code != codes.Error {
			panic("wrong span status on span 0")
		}
		if stubs[0][1].Status.Code != codes.Error {
			panic("wrong span status on span 1")
		}
	}, 1)
}
