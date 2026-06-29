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
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type greeterService struct {
	UnimplementedHelloGrpcServer
}

func (s *greeterService) Hello(ctx context.Context, req *Req) (*Resp, error) {
	name := req.GetName()
	if name == "" {
		name = "World"
	}
	return &Resp{Message: fmt.Sprintf("Hello %s!", name)}, nil
}

func (s *greeterService) StreamHello(req *Req, stream HelloGrpc_StreamHelloServer) error {
	name := req.GetName()
	if name == "" {
		name = "World"
	}
	for i := 1; i <= 3; i++ {
		if err := stream.Send(&Resp{Message: fmt.Sprintf("[%d] Hello %s!", i, name)}); err != nil {
			return err
		}
	}
	return nil
}

func startServer() {
	lis, err := net.Listen("tcp", "0.0.0.0:9003")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterHelloGrpcServer(s, &greeterService{})
	log.Println("gRPC server listening on :9003")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func callUnary(ctx context.Context) {
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := NewHelloGrpcClient(conn)
	resp, err := client.Hello(ctx, &Req{Name: "LoonGSuite"})
	if err != nil {
		log.Fatalf("unary call failed: %v", err)
	}
	fmt.Printf("Unary response: %s\n", resp.GetMessage())
}

func callServerStream(ctx context.Context) {
	conn, err := grpc.NewClient("localhost:9003", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := NewHelloGrpcClient(conn)
	stream, err := client.StreamHello(ctx, &Req{Name: "LoonGSuite"})
	if err != nil {
		log.Fatalf("stream call failed: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("stream recv failed: %v", err)
		}
		fmt.Printf("Stream response: %s\n", resp.GetMessage())
	}
}

func main() {
	go startServer()
	time.Sleep(1 * time.Second)

	ctx := context.Background()

	fmt.Println("=== Unary RPC ===")
	callUnary(ctx)

	fmt.Println("\n=== Server Streaming RPC ===")
	callServerStream(ctx)

	fmt.Println("\nDone!")
}
