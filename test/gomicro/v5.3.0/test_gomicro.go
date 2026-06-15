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
	"fmt"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	example "github.com/go-micro/examples/server/proto/example"
	micro "go-micro.dev/v5"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/metadata"
)

func call(i int, c client.Client) {
	// Create new request to service go.micro.srv.example, method Example.Call
	req := c.NewRequest("go.micro.srv.example", "Example.Call", &example.Request{
		Name: "John",
	})

	// create context with metadata
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

	rsp := &example.Response{}

	// Call service
	if err := c.Call(ctx, req, rsp); err != nil {
		fmt.Println("call err: ", err, rsp)
		return
	}

	fmt.Println("Call:", i, "rsp:", rsp.Msg)

}

func requestServer() {
	service := micro.NewService()
	service.Init()

	fmt.Println("\n--- Call example ---")
	call(10, service.Client())
}

func main() {
	// starter server
	go setupHttp()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	requestServer()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "Example.Call", "Example.Call", "Example.Call", "http", "", "tcp", "ipv4", "", "go.micro.srv.example", 200, 0, 0)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "Example.Call Example.Call", "Example.Call", "http", "tcp", "ipv4", "", "go.micro.srv.example", "Go-http-client/1.1", "http", "Example.Call", "", "Example.Call", 200)
	}, 1)
}
