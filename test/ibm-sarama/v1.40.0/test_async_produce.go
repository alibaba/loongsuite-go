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

	// create async producer
	producer, err := createAsyncProducer()
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

	// send message via async producer
	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Value: sarama.ByteEncoder("hello async sarama"),
	}
	producer.Input() <- msg

	// wait for success or error
	select {
	case successMsg := <-producer.Successes():
		fmt.Printf("produced message to partition %d at offset %d\n", successMsg.Partition, successMsg.Offset)
	case errMsg := <-producer.Errors():
		panic(fmt.Sprintf("failed to produce message: %v", errMsg.Err))
	case <-time.After(20 * time.Second):
		panic("timeout waiting for produce result")
	}

	// consume the message
	timer := time.NewTimer(30 * time.Second)
	select {
	case message := <-partitionConsumer.Messages():
		fmt.Printf("consumed message: %s\n", string(message.Value))
	case <-timer.C:
		panic("timeout waiting for message")
	}

	time.Sleep(2 * time.Second)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyMQPublishAttributes(stubs[0][0], "", "", "", "publish", topicName, "kafka")
		verifier.VerifyMQConsumeAttributes(stubs[0][1], "", "", "", "receive", topicName, "kafka")

		if stubs[0][1].Parent.TraceID().String() != stubs[0][0].SpanContext.TraceID().String() {
			panic("consumer span should share trace ID with producer span")
		}
	}, 1)
}
