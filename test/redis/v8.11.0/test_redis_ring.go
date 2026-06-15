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
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"shard1": "localhost:" + os.Getenv("REDIS_PORT"),
		},
	})
	_, err := rdb.HSet(ctx, "a", map[string]string{
		"a": "b",
	}).Result()
	if err != nil {
		panic(err)
	}
	val := rdb.HVals(ctx, "a").Val()
	fmt.Printf("%v\n", val)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "hset", "redis", "shard1", "hset a a b", "hset", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "hvals", "redis", "shard1", "hvals a", "hvals", "", nil)
	}, 3)
}
