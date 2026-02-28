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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net"
)

type service struct {
	HelloGrpcServer
}

func (s *service) Hello(ctx context.Context, req *Req) (*Resp, error) {
	if req.Error {
		return &Resp{Message: "Error Grpc"}, errors.New("error Grpc")
	}
	return &Resp{Message: "Hello Gprc"}, nil
}

func (s *service) StreamHello(req *Req, res HelloGrpc_StreamHelloServer) error {
	if req.Error {
		return errors.New("error Grpc")
	}
	for i := 0; i < 2; i++ {
		err := res.Send(&Resp{Message: "Hello Gprc"})
		if err != nil {
			return err
		}
	}
	return nil
}

func setupGRPC() {
	lis, err := net.Listen("tcp", "0.0.0.0:9003")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterHelloGrpcServer(s, &service{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func sendReq(ctx context.Context) string {
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := NewHelloGrpcClient(conn)
	resp, err := c.Hello(ctx, &Req{})
	if err != nil {
		panic(err)
	}
	return resp.Message
}

func sendErrReq(ctx context.Context) string {
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := NewHelloGrpcClient(conn)
	resp, err := c.Hello(ctx, &Req{Error: true})
	if err != nil {
		log.Printf("error %v\n", err)
	}
	if resp != nil {
		return resp.Message
	}
	return "error resp"
}

func sendStreamReq(ctx context.Context) string {
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := NewHelloGrpcClient(conn)
	stream, err := c.StreamHello(ctx, &Req{})
	if err != nil {
		panic(err)
	}
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
	return "stream success"
}
