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

	"github.com/Shopify/sarama"
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

	// send multiple messages via sync producer (individual calls)
	msgCount := 3
	for i := 0; i < msgCount; i++ {
		msg := &sarama.ProducerMessage{
			Topic: topicName,
			Value: sarama.ByteEncoder(fmt.Sprintf("batch message %d", i+1)),
		}
		_, _, err = producer.SendMessage(msg)
		if err != nil {
			panic(fmt.Sprintf("failed to send message %d: %v", i+1, err))
		}
	}
	fmt.Println("produced multiple messages successfully")

	// consume all messages
	for i := 0; i < msgCount; i++ {
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
		// Each message should produce a trace with publish + receive spans
		for i, trace := range stubs {
			verifier.VerifyMQPublishAttributes(trace[0], "", "", "", "publish", topicName, "kafka")
			verifier.VerifyMQConsumeAttributes(trace[1], "", "", "", "receive", topicName, "kafka")

			if trace[1].Parent.TraceID().String() != trace[0].SpanContext.TraceID().String() {
				panic(fmt.Sprintf("trace %d: consumer span should share trace ID with producer span", i))
			}
		}
	}, 3)
}
