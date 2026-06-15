package main

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"time"
)

func main() {
	// starter server
	go setupGRPC()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	sendErrReq(context.Background())
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyRpcClientAttributes(stubs[0][0], "/HelloGrpc/Hello", "grpc", "/HelloGrpc", "Hello")
		verifier.VerifyRpcServerAttributes(stubs[0][1], "/HelloGrpc/Hello", "grpc", "/HelloGrpc", "Hello")
		if stubs[0][0].Status.Code != codes.Error {
			panic("wrong span status on span 0")
		}
		if stubs[0][1].Status.Code != codes.Error {
			panic("wrong span status on span 1")
		}
	}, 1)

}
