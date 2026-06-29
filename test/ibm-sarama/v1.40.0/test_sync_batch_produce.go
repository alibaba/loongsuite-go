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
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	if err := createTopic(); err != nil {
		panic(fmt.Sprintf("create topic error: %v", err))
	}

	// create sync producer
	producer, err := createSyncProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	// create partition consumer
	consumer, err := createPartitionConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topicName, 0, sarama.OffsetNewest)
	if err != nil {
		panic(fmt.Sprintf("failed to create partition consumer: %v", err))
	}
	defer partitionConsumer.Close()

	// send batch messages via sync producer
	msgs := []*sarama.ProducerMessage{
		{
			Topic: topicName,
			Value: sarama.ByteEncoder("batch message 1"),
		},
		{
			Topic: topicName,
			Value: sarama.ByteEncoder("batch message 2"),
		},
	}

	err = producer.SendMessages(msgs)
	if err != nil {
		panic(fmt.Sprintf("failed to send batch messages: %v", err))
	}
	fmt.Println("produced batch messages successfully")

	// consume both messages
	for i := 0; i < 2; i++ {
		timer := time.NewTimer(30 * time.Second)
		select {
		case message := <-partitionConsumer.Messages():
			fmt.Printf("consumed message %d: %s\n", i+1, string(message.Value))
		case <-timer.C:
			panic(fmt.Sprintf("timeout waiting for message %d", i+1))
		}
	}

	time.Sleep(2 * time.Second)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		// Expect publish and receive spans for each message
		// The batch produce should create spans for each message
		for _, stub := range stubs[0] {
			fmt.Printf("span: %s\n", stub.Name)
		}

		// Verify that we have publish and receive spans
		publishCount := 0
		receiveCount := 0
		for _, stub := range stubs[0] {
			for _, attr := range stub.Attributes {
				if string(attr.Key) == "messaging.operation" {
					if attr.Value.AsString() == "publish" {
						publishCount++
					} else if attr.Value.AsString() == "receive" {
						receiveCount++
					}
				}
			}
		}

		if publishCount < 2 {
			panic(fmt.Sprintf("expected at least 2 publish spans, got %d", publishCount))
		}
		if receiveCount < 2 {
			panic(fmt.Sprintf("expected at least 2 receive spans, got %d", receiveCount))
		}
	}, 1)
}
