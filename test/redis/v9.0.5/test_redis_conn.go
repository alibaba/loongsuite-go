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
	"time"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + os.Getenv("REDIS_PORT"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	cn := rdb.Conn()
	defer cn.Close()

	if err := cn.ClientSetName(ctx, "myclient").Err(); err != nil {
		panic(err)
	}

	name, err := cn.ClientGetName(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("client name", name)

	_, err = cn.Set(ctx, "a", "b", 5*time.Second).Result()
	if err != nil {
		panic(err)
	}
	_, err = cn.Get(ctx, "a").Result()
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "client", "redis", "localhost", "client setname myclient", "client", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "client", "redis", "localhost", "client getname", "client", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "set", "redis", "localhost", "set a b ex 5", "set", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "get", "redis", "localhost", "get a", "get", "", nil)
	}, 4)
}
