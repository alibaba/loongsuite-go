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
	"errors"
	"fmt"
	"os"

	"github.com/IBM/sarama"
)

const (
	topicName = "test-sarama-topic"
)

var kafkaVersion = sarama.V2_1_0_0

func getKafkaAddress() string {
	if addr := os.Getenv("KAFKA_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:9092"
}

func createSyncProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	return sarama.NewSyncProducer([]string{getKafkaAddress()}, config)
}

func createAsyncProducer() (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	return sarama.NewAsyncProducer([]string{getKafkaAddress()}, config)
}

func createPartitionConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	return sarama.NewConsumer([]string{getKafkaAddress()}, config)
}

func createTopic() error {
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	admin, err := sarama.NewClusterAdmin([]string{getKafkaAddress()}, config)
	if err != nil {
		return err
	}
	defer admin.Close()
	err = admin.CreateTopic(topicName, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		var topicError *sarama.TopicError
		if errors.As(err, &topicError) && topicError.Err == sarama.ErrTopicAlreadyExists {
			fmt.Println("topic already exists")
			return nil
		}
		return err
	}
	return nil
}
