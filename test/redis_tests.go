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

package test

import (
	"context"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const redisv9_dependency_name = "github.com/redis/go-redis/v9"
const redisv8_dependency_name = "github.com/go-redis/redis/v8"
const redis_module_name = "redis"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("redis-9.0.5-executing-commands-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestExecutingCommands),
		NewGeneralTestCase("redis-9.0.5-executing-unsupported-commands-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestExecutingUnsupportedCommands),
		NewGeneralTestCase("redis-9.0.5-redis-conn-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestRedisConn),
		NewGeneralTestCase("redis-9.0.5-ring-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestRedisRing),
		NewGeneralTestCase("redis-9.0.5-transactions-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestRedisTransactions),
		NewGeneralTestCase("redis-9.0.5-universal-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestRedisUniversal),
		NewGeneralTestCase("redis-8.11.0-executing-commands-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8ExecutingCommands),
		NewGeneralTestCase("redis-8.11.0-executing-unsupported-commands-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8ExecutingUnsupportedCommands),
		NewGeneralTestCase("redis-8.11.0-redis-conn-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8RedisConn),
		NewGeneralTestCase("redis-8.11.0-ring-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8RedisRing),
		NewGeneralTestCase("redis-8.11.0-transactions-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8RedisTransactions),
		NewGeneralTestCase("redis-8.11.0-universal-test", redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", TestV8RedisUniversal),
		NewMuzzleTestCase("redis-8.11.0-muzzle", redisv8_dependency_name, redis_module_name, "v8.11.0", "v8.11.5", "1.18", "", []string{"go", "build", "test_executing_commands.go"}),
		NewMuzzleTestCase("redis-9.0.5-muzzle", redisv9_dependency_name, redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", []string{"go", "build", "test_executing_commands.go"}),
		NewLatestDepthTestCase("redis-9.0.5-executing-commands-latestDepth", redisv9_dependency_name, redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestExecutingCommands))
}

func TestExecutingCommands(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_executing_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_commands", env...)
}

func TestExecutingUnsupportedCommands(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_executing_unsupported_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_unsupported_commands", env...)
}

func TestRedisConn(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_redis_conn.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_conn", env...)
}

func TestRedisRing(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_redis_ring.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_ring", env...)
}

func TestRedisTransactions(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_redis_transactions.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_transactions", env...)
}

func TestRedisUniversal(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_universal_client.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_universal_client", env...)
}

func TestV8ExecutingCommands(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_executing_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_commands", env...)
}

func TestV8ExecutingUnsupportedCommands(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_executing_unsupported_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_unsupported_commands", env...)
}

func TestV8RedisConn(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_redis_conn.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_conn", env...)
}

func TestV8RedisRing(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_redis_ring.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_ring", env...)
}

func TestV8RedisTransactions(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_redis_transactions.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_redis_transactions", env...)
}

func TestV8RedisUniversal(t *testing.T, env ...string) {
	_, redisPort := initRedisContainer()
	UseApp("redis/v8.11.0")
	RunGoBuild(t, "go", "build", "test_universal_client.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_universal_client", env...)
}

func initRedisContainer() (testcontainers.Container, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	port, err := redisC.MappedPort(context.Background(), "6379")
	if err != nil {
		panic(err)
	}
	return redisC, port
}
