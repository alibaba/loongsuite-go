// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

const streadwayAMQPDependencyName = "github.com/streadway/amqp"
const streadwayAMQPModuleName = "streadway-amqp"

func init() {
	tc1 := NewGeneralTestCase("streadway-amqp-cascading-1.0.0-test", streadwayAMQPModuleName, "v1.0.0", "", "1.24", "", TestStreadwayAMQPCascading)
	tc2 := NewGeneralTestCase("streadway-amqp-no-cascading-1.0.0-test", streadwayAMQPModuleName, "v1.0.0", "", "1.24", "", TestStreadwayAMQPNoCascading)
	tc3 := NewMuzzleTestCase("streadway-amqp-muzzle-1.0.0-test", streadwayAMQPDependencyName, streadwayAMQPModuleName, "v1.0.0", "", "1.24", "", []string{"go", "build", "test_mq_cascading.go", "base.go"})

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

func TestStreadwayAMQPCascading(t *testing.T, env ...string) {
	rabbitC, port := initStreadwayAMQPContainer()
	t.Cleanup(func() {
		if err := rabbitC.Terminate(context.Background()); err != nil {
			t.Errorf("failed to terminate rabbitmq container: %v", err)
		}
	})

	UseApp("streadway-amqp/v1.0.0")
	RunGoBuild(t, "go", "build", "test_mq_cascading.go", "base.go")
	env = append(env, "RabbitMQ_PORT="+port.Port())
	RunApp(t, "test_mq_cascading", env...)
}

func TestStreadwayAMQPNoCascading(t *testing.T, env ...string) {
	rabbitC, port := initStreadwayAMQPContainer()
	t.Cleanup(func() {
		if err := rabbitC.Terminate(context.Background()); err != nil {
			t.Errorf("failed to terminate rabbitmq container: %v", err)
		}
	})

	UseApp("streadway-amqp/v1.0.0")
	RunGoBuild(t, "go", "build", "test_mq_no_cascading.go", "base.go")
	env = append(env, "RabbitMQ_PORT="+port.Port())
	RunApp(t, "test_mq_no_cascading", env...)
}

func initStreadwayAMQPContainer() (testcontainers.Container, nat.Port) {
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
