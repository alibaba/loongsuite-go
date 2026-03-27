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

const asynqDependencyName = "github.com/hibiken/asynq"
const asynqModuleName = "asynq"

func init() {
	tc1 := NewGeneralTestCase("asynq-enqueue-process-0.23.0-test", asynqModuleName, "0.23.0", "0.23.0", "1.21", "", TestAsynqEnqueueProcessV023)
	tc2 := NewGeneralTestCase("asynq-enqueue-process-0.26.0-test", asynqModuleName, "0.26.0", "0.26.0", "1.24", "", TestAsynqEnqueueProcessV026)
	tc3 := NewMuzzleTestCase("asynq-muzzle-0.23.0-test", asynqDependencyName, asynqModuleName, "0.23.0", "0.23.0", "1.21", "", []string{"go", "build", "test_enqueue_process.go"})

	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
	if tc3 != nil {
		TestCases = append(TestCases, tc3)
	}
}

func TestAsynqEnqueueProcessV023(t *testing.T, env ...string) {
	_, redisPort := initAsynqRedisContainer()
	UseApp("asynq/v0.23.0")
	RunGoBuild(t, "go", "build", "test_enqueue_process.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_enqueue_process", env...)
}

func TestAsynqEnqueueProcessV026(t *testing.T, env ...string) {
	_, redisPort := initAsynqRedisContainer()
	UseApp("asynq/v0.26.0")
	RunGoBuild(t, "go", "build", "test_enqueue_process.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_enqueue_process", env...)
}

func initAsynqRedisContainer() (testcontainers.Container, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "registry.cn-hangzhou.aliyuncs.com/private-mesh/hellob:redis",
		ReaperImage:  "registry.cn-hangzhou.aliyuncs.com/private-mesh/hellob:redis",
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
