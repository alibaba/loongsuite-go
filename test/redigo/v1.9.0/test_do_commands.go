package main

import (
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/gomodule/redigo/redis"
)

func main() {
	c, err := redis.Dial("tcp", "localhost:"+os.Getenv("REDIS_PORT"))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	c.Do("SET", "foo", "bar")
	c.Do("GET", "foo")

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "SET", "redis", "localhost", "SET foo bar", "SET", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "GET", "redis", "localhost", "GET foo", "GET", "", nil)
	}, 2)
}
