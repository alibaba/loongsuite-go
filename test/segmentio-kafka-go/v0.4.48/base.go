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

package main

import (
	"os"

	"github.com/segmentio/kafka-go"
)

const (
	topicName = "test-topic"
	groupName = "test-group"
)

// getKafkaAddress returns Kafka broker address from environment or default
func getKafkaAddress() string {
	if addr := os.Getenv("KAFKA_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:9092" // Default Kafka address
}

// initProducer creates a new Kafka producer configured for our test topic
func initProducer() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(getKafkaAddress()),
		Topic:    topicName,
		Balancer: &kafka.LeastBytes{},
		Async:    false, // Synchronous writes for testing
	}
}

// initConsumer creates a new Kafka consumer configured for our test topic
func initConsumer() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{getKafkaAddress()},
		Topic:    topicName,
		GroupID:  groupName,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
