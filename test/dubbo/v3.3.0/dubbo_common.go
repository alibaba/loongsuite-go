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
	"errors"

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/client"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"github.com/dubbogo/gost/log/logger"
)

type GreetTripleServer struct {
}

func (srv *GreetTripleServer) Greet(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
	if req.Error {
		return nil, errors.New("error triple")
	}
	return &GreetResponse{Greeting: "Hello" + req.Name}, nil
}

func setupDubbo() {
	ins, err := dubbo.NewInstance(
		dubbo.WithName("dubbo_test_server"),
		dubbo.WithProtocol(
			protocol.WithTriple(),
			protocol.WithPort(20000),
		),
	)
	if err != nil {
		panic(err)
	}
	srv, err := ins.NewServer()
	if err != nil {
		panic(err)
	}
	if err := RegisterGreetServiceHandler(srv, &GreetTripleServer{}); err != nil {
		panic(err)
	}

	if err := srv.Serve(); err != nil {
		logger.Error(err)
	}
}

func sendBasicDubboReq(ctx context.Context) {
	instance, err := dubbo.NewInstance(
		dubbo.WithName("dubbo_test_client"),
		dubbo.WithProtocol(
			protocol.WithTriple()),
	)
	if err != nil {
		panic(err)
	}

	cli, err := instance.NewClient(
		client.WithClientURL("tri://127.0.0.1:20000"),
	)
	if err != nil {
		panic(err)
	}

	svc, err := NewGreetService(cli)
	if err != nil {
		panic(err)
	}

	resp, err := svc.Greet(ctx, &GreetRequest{Name: "Alibaba"})
	if err != nil {
		panic(err)
	}
	logger.Infof("Greet response: %s", resp)
}

func sendErrDubboReq(ctx context.Context) {
	instance, err := dubbo.NewInstance(
		dubbo.WithName("dubbo_test_client"),
		dubbo.WithProtocol(
			protocol.WithTriple()),
	)
	if err != nil {
		panic(err)
	}

	cli, err := instance.NewClient(
		client.WithClientURL("tri://127.0.0.1:20000"),
	)
	if err != nil {
		panic(err)
	}

	svc, err := NewGreetService(cli)
	if err != nil {
		panic(err)
	}

	_, err = svc.Greet(ctx, &GreetRequest{Error: true})
	if err != nil {
		logger.Infof("err %v\n", err)
	}
}
