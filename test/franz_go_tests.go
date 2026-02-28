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
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const franzModuleName = "franzkafka"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("franz-go-1.18.0-test", franzModuleName, "v1.18.0", "", "1.22.0", "", TestFranzGoBasic),
	)
}

func TestFranzGoBasic(t *testing.T, env ...string) {
	containers := initFranzKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())

	UseApp("franzkafka/v1.18.0")
	RunGoBuild(t, "go", "build", "test_franz.go")
	env = append(env, "FRANZ_KAFKA_PORT="+containers.KafkaPort)
	RunApp(t, "test_franz", env...)
}

type FranzKafkaContainers struct {
	KafkaContainer testcontainers.Container
	KafkaPort      string
	network        testcontainers.Network
}

func (c *FranzKafkaContainers) CleanupContainers(ctx context.Context) error {
	var errs []error
	if c.KafkaContainer != nil {
		if err := c.KafkaContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate Kafka container: %w", err))
		}
	}
	if c.network != nil {
		if err := c.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove network: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

func initFranzKafkaContainer(t *testing.T) *FranzKafkaContainers {
	ctx := context.Background()

	testNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	containers := &FranzKafkaContainers{network: testNetwork}

	kafkaReq := testcontainers.ContainerRequest{
		Image:        "registry.cn-hangzhou.aliyuncs.com/private-mesh/hellob:kafka-370",
		ExposedPorts: []string{"9092/tcp"},
		Env: map[string]string{
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_LISTENERS":                                "PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://localhost:9092",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9093",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "3000",
			"KAFKA_NUM_PARTITIONS":                           "1",
			"KAFKA_LOG_DIRS":                                 "/tmp/kraft-combined-logs",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				"9092/tcp": []nat.PortBinding{{
					HostIP:   "0.0.0.0",
					HostPort: "9092",
				}},
			}
		},
		WaitingFor: wait.ForLog("Kafka Server started").WithStartupTimeout(60 * time.Second),
		Networks:   []string{testNetwork.Name},
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Kafka container: %v", err)
	}

	time.Sleep(10 * time.Second)

	containers.KafkaContainer = kafkaC
	containers.KafkaPort = "9092"
	return containers
}
