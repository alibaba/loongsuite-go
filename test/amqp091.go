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

const rabbitmq_dependency_name = "https://github.com/rabbitmq/amqp091-go"
const rabbitmq_module_name = "amqp091"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("rabbitmq_cascading-1.10.0-test", rabbitmq_module_name, "1.10.0", "1.10.0", "1.22.0", "", TestRabbitMQCascading),
		NewGeneralTestCase("rabbitmq_no_cascading-1.10.0-test", rabbitmq_module_name, "1.10.0", "1.10.0", "1.22.0", "", TestRabbitMQNOCascading),
	)

}
func TestRabbitMQCascading(t *testing.T, env ...string) {
	_, port := initRabbitMQContainer()
	UseApp("amqp091/v1.10.0")
	RunGoBuild(t, "go", "build", "test_mq_cascading.go", "base.go")
	env = append(env, "RabbitMQ_PORT="+port.Port())
	RunApp(t, "test_mq_cascading", env...)
}
func TestRabbitMQNOCascading(t *testing.T, env ...string) {
	_, port := initRabbitMQContainer()
	UseApp("amqp091/v1.10.0")
	RunGoBuild(t, "go", "build", "test_mq_no_cascading.go", "base.go")
	env = append(env, "RabbitMQ_PORT="+port.Port())
	RunApp(t, "test_mq_no_cascading", env...)
}
func initRabbitMQContainer() (testcontainers.Container, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:4.0.7-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbitC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	port, err := rabbitC.MappedPort(context.Background(), "5672")
	if err != nil {
		panic(err)
	}
	return rabbitC, port
}
