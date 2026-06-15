// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

const rueidisgo_dependency_name = "github.com/redis/rueidis"
const rueidis_module_name = "rueidis"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("rueidis-1.0.30-basic-test", rueidis_module_name, "v1.0.30", "", "1.18", "", TestRueidisExecutingCommands))
}

func TestRueidisExecutingCommands(t *testing.T, env ...string) {
	_, redisPort := initRueidisContainer()
	UseApp("rueidis/v1.0.30")
	RunGoBuild(t, "go", "build", "test_basic.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_basic", env...)
}

func initRueidisContainer() (testcontainers.Container, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "registry.cn-hangzhou.aliyuncs.com/private-mesh/hellob:redis",
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
