// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const kafkaModuleName = "segmentio-kafka-go"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("segmentio-kafka-go-basic-test", kafkaModuleName, "0.4.0", "", "1.18.0", "", TestBasicKafka),
	)
}

func TestBasicKafka(t *testing.T, env ...string) {
	containers := initKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())
	UseApp("segmentio-kafka-go/v0.4.48")
	RunGoBuild(t, "go", "build", "test_kafka_basic.go", "base.go")
	env = append(env, "KAFKA_ADDR="+containers.KafkaAddress)
	RunApp(t, "test_kafka_basic", env...)
}

// KafkaContainers encapsulates Kafka and Zookeeper containers for unified management
type KafkaContainers struct {
	ZookeeperContainer testcontainers.Container
	KafkaContainer     testcontainers.Container
	ZookeeperAddress   string
	KafkaAddress       string
	network            testcontainers.Network
}

// CleanupContainers Cleanup terminates all containers and network resources
func (k *KafkaContainers) CleanupContainers(ctx context.Context) error {
	var errs []error

	if k.KafkaContainer != nil {
		if err := k.KafkaContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate Kafka container: %w", err))
		}
	}

	if k.ZookeeperContainer != nil {
		if err := k.ZookeeperContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate Zookeeper container: %w", err))
		}
	}

	if k.network != nil {
		if err := k.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove network: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

// initKafkaContainer initializes a complete Kafka testing environment
func initKafkaContainer(t *testing.T) *KafkaContainers {
	ctx := context.Background()

	// Create a dedicated network for the containers
	testNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	containers := &KafkaContainers{network: testNetwork}

	// Start Zookeeper container
	zookeeperReq := testcontainers.ContainerRequest{
		Image:        "wurstmeister/zookeeper:latest",
		ExposedPorts: []string{"2181/tcp"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
		},
		WaitingFor: wait.ForLog("binding to port").WithStartupTimeout(30 * time.Second),
		Networks:   []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"zookeeper"},
		},
	}

	zookeeperC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: zookeeperReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Zookeeper container: %v", err)
	}

	hostIP := "127.0.0.1" // For test environment access
	zkAddr := net.JoinHostPort(hostIP, "2181")

	// Start Kafka container
	kafkaReq := testcontainers.ContainerRequest{
		Image:        "wurstmeister/kafka:2.12-2.4.1",
		ExposedPorts: []string{"9092/tcp"},
		Env: map[string]string{
			"KAFKA_ADVERTISED_LISTENERS":      fmt.Sprintf("PLAINTEXT://%s:9092", hostIP),
			"KAFKA_LISTENERS":                 "PLAINTEXT://0.0.0.0:9092",
			"KAFKA_ZOOKEEPER_CONNECT":         "zookeeper:2181",
			"KAFKA_CREATE_TOPICS":             "test-topic:1:1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "true",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				"9092/tcp": []nat.PortBinding{{
					HostIP:   "0.0.0.0",
					HostPort: "9092",
				}},
			}
		},
		WaitingFor: wait.ForLog("started (kafka.server.KafkaServer)").WithStartupTimeout(60 * time.Second),
		Networks:   []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"kafka"},
		},
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaReq,
		Started:          true,
	})
	if err != nil {
		_ = zookeeperC.Terminate(ctx)
		t.Fatalf("Failed to start Kafka container: %v", err)
	}

	kafkaAddr := net.JoinHostPort(hostIP, "9092")

	// Wait for Kafka to be fully ready
	time.Sleep(5 * time.Second)

	containers.ZookeeperContainer = zookeeperC
	containers.KafkaContainer = kafkaC
	containers.ZookeeperAddress = zkAddr
	containers.KafkaAddress = kafkaAddr

	return containers
}
