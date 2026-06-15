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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var kafkaBroker = "localhost:" + os.Getenv("FRANZ_KAFKA_PORT")

const (
	topic         = "franz-demo-topic"
	consumerGroup = "my-franz-group"
)

func main() {
	opts := []kgo.Opt{
		kgo.SeedBrokers(kafkaBroker),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topic),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	}

	log.Println("Connecting to Kafka...")
	client, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer client.Close()
	log.Println("Successfully connected!")

	log.Println("--- Admin: Ensuring topic exists ---")
	createTopic(client, topic)

	log.Println("--- Start Producing ---")
	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		message := fmt.Sprintf("Hello Franz-Go! Message #%d", i)
		record := &kgo.Record{Topic: topic, Value: []byte(message)}

		client.Produce(context.Background(), record, func(r *kgo.Record, err error) {
			defer wg.Done()
			if err != nil {
				log.Printf("Failed to produce message: %v\n", err)
			} else {
				log.Printf("Produced message to topic '%s', partition %d, offset %d\n",
					r.Topic, r.Partition, r.Offset)
			}
		})
	}

	log.Println("Waiting for all messages to be produced...")
	wg.Wait()
	log.Println("--- Finished Producing ---")

	log.Println("--- Start Consuming ---")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for {
		fetches := client.PollFetches(ctx)

		if fetches.IsClientClosed() || ctx.Err() != nil {
			log.Println("Consumer loop is exiting...")
			break
		}

		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("Fetch error for topic %s, partition %d: %v\n", t, p, err)
		})

		fetches.EachRecord(func(r *kgo.Record) {
			log.Printf("Consumed message: '%s' from topic '%s', partition %d, offset %d\n",
				string(r.Value), r.Topic, r.Partition, r.Offset)
		})
	}

	log.Println("--- Finished Consuming ---")
	time.Sleep(5 * time.Second)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		xx, _ := json.Marshal(stubs)
		fmt.Println(string(xx))
		verifier.VerifyMQPublishAttributes(stubs[0][0], "", "", "", "publish", topic, "kafka")
		verifier.VerifyMQConsumeAttributes(stubs[0][1], "", "", "", "process", topic, "kafka")
	}, 1)
	log.Println("--- Finished Verifying Traces ---")
}

func createTopic(client *kgo.Client, topicName string) {
	adminClient := kadm.NewClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Creating topic '%s'...", topicName)
	responses, err := adminClient.CreateTopics(ctx, 3, 1, nil, topicName)
	if err != nil {
		log.Printf("Error when creating topic: %v", err)
		return
	}
	for _, response := range responses {
		if response.Err != nil {
			log.Printf("Error creating topic '%s': %v", response.Topic, response.Err)
		} else {
			log.Printf("Successfully created topic '%s' with ID: %s", response.Topic, response.ID)
		}
	}
}
