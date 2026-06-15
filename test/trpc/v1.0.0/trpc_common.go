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
	"fmt"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
)

type GreeterServer struct{}

func (s *GreeterServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	value := fmt.Sprintf("Hello, %s!", req.Name)
	return &HelloResponse{
		Message: value,
	}, nil
}

func setupTrpcServer() {
	s := trpc.NewServer()
	RegisterGreeterService(s, &GreeterServer{})
	if err := s.Serve(); err != nil {
		log.Fatal(err)
	}
}

func clientSendReq() {
	_ = trpc.NewServer()

	proxy := NewGreeterClientProxy(client.WithTarget("ip://127.0.0.1:8000"))

	req := &HelloRequest{
		Name: "tRPC",
	}
	_, err := proxy.SayHello(context.Background(), req)

	if err != nil {
		fmt.Printf("call SayHello failed: %v\n", err)
		return
	}
}
