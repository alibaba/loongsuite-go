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
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"
)

type MyHash struct {
	Key1 string `redis:"key1"`
	Key2 int    `redis:"key2"`
}

func main() {
	ctx := context.Background()
	rdb := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"shard1": "localhost:" + os.Getenv("REDIS_PORT"),
		},
		Password: "Hello1234",
	})
	_, err := rdb.HSet(ctx, "a", MyHash{
		Key1: "1",
		Key2: 2,
	}).Result()
	if err != nil {
		panic(err)
	}
	val := rdb.HVals(ctx, "a").Val()
	fmt.Printf("%v\n", val)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "hset", "redis", "localhost", "hset a key1 1 key2 2", "hset", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "hvals", "redis", "localhost", "hvals a", "hvals", "", nil)
	}, 2)
}
