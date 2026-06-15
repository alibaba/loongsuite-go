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

const redigo_dependency_name = "github.com/gomodule/redigo"
const redigo_module_name = "redigo"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("redigo-1.9.0-executing-commands-test", redigo_module_name, "v1.9.0", "", "1.18", "", TestRedigoExecutingCommands),
		NewGeneralTestCase("redigo-1.9.0-do-commands-test", redigo_module_name, "v1.9.0", "", "1.18", "", TestRedigoDoCommands),
		NewGeneralTestCase("redigo-1.9.0-unsupported-commands-test", redigo_module_name, "v1.9.0", "", "1.18", "", TestRedigoUnsupportedCommands),
		NewGeneralTestCase("redigo-1.9.0-transaction-test", redigo_module_name, "v1.9.0", "", "1.18", "", TestRedigoTransactions),
		NewMuzzleTestCase("redigo-muzzle-test", redigo_dependency_name, redigo_module_name, "v1.9.0", "", "1.18", "", []string{"go", "build", "test_do_commands.go"}),
		NewLatestDepthTestCase("redigo-latest-depth-test", redigo_dependency_name, redigo_module_name, "v1.9.0", "", "1.18", "", TestRedigoDoCommands),
	)
}

func TestRedigoExecutingCommands(t *testing.T, env ...string) {
	_, redisPort := initRedigoContainer()
	UseApp("redigo/v1.9.0")
	RunGoBuild(t, "go", "build", "test_executing_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_commands", env...)
}

func TestRedigoDoCommands(t *testing.T, env ...string) {
	_, redisPort := initRedigoContainer()
	UseApp("redigo/v1.9.0")
	RunGoBuild(t, "go", "build", "test_do_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_do_commands", env...)
}

func TestRedigoUnsupportedCommands(t *testing.T, env ...string) {
	_, redisPort := initRedigoContainer()
	UseApp("redigo/v1.9.0")
	RunGoBuild(t, "go", "build", "test_unsupported_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_unsupported_commands", env...)
}

func TestRedigoTransactions(t *testing.T, env ...string) {
	_, redisPort := initRedigoContainer()
	UseApp("redigo/v1.9.0")
	RunGoBuild(t, "go", "build", "test_transaction.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_transaction", env...)
}

func initRedigoContainer() (testcontainers.Container, nat.Port) {
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
