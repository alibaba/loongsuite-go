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

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// The first Config call comes from OnEnter. Blocking the second call parks
// the actual constructor while its goroutine-local suppression is active.
type blockedClient struct {
	sarama.Client
	calls   int
	entered chan struct{}
	release chan struct{}
}

func (c *blockedClient) Config() *sarama.Config {
	c.calls++
	if c.calls == 2 {
		close(c.entered)
		select {
		case <-c.release:
		case <-time.After(20 * time.Second):
			panic("timeout waiting to release sync constructor")
		}
	}
	return c.Client.Config()
}

func newClient() sarama.Client {
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	client, err := sarama.NewClient([]string{getKafkaAddress()}, config)
	if err != nil {
		panic(err)
	}
	return client
}

func sendSync(producer sarama.SyncProducer) {
	defer producer.Close()
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicName, Value: sarama.StringEncoder("concurrent sync"),
	})
	if err != nil {
		panic(err)
	}
}

func sendAsync() {
	producer, err := createAsyncProducer()
	if err != nil {
		panic(err)
	}
	defer producer.Close()
	producer.Input() <- &sarama.ProducerMessage{
		Topic: topicName, Value: sarama.StringEncoder("independent async"),
	}
	select {
	case <-producer.Successes():
	case err := <-producer.Errors():
		panic(err)
	case <-time.After(20 * time.Second):
		panic("timeout waiting for async publish")
	}
}

func main() {
	const producerCount = 16
	if err := createTopic(); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(producerCount)
	for i := range producerCount {
		go func() {
			defer wg.Done()
			var producer sarama.SyncProducer
			var err error
			if i%2 == 0 {
				producer, err = createSyncProducer()
			} else {
				client := newClient()
				defer client.Close()
				producer, err = sarama.NewSyncProducerFromClient(client)
			}
			if err != nil {
				panic(err)
			}
			sendSync(producer)
		}()
	}
	wg.Wait()

	// Deterministically overlap an independent async producer with a sync
	// constructor. A process-wide suppression flag loses this publish span.
	client := &blockedClient{
		Client: newClient(), entered: make(chan struct{}), release: make(chan struct{}),
	}
	defer client.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		producer, err := sarama.NewSyncProducerFromClient(client)
		if err != nil {
			panic(err)
		}
		sendSync(producer)
	}()
	select {
	case <-client.entered:
	case <-time.After(20 * time.Second):
		panic("timeout waiting for sync constructor")
	}
	sendAsync()
	close(client.release)
	<-done

	// A panic in OnEnter must not suppress later producers on this goroutine.
	func() {
		defer func() { _ = recover() }()
		_, _ = sarama.NewSyncProducerFromClient(nil)
	}()
	sendAsync()

	verifier.WaitAndAssertTraces(func(traces []tracetest.SpanStubs) {
		if len(traces) != producerCount+3 {
			panic(fmt.Sprintf("got %d publish traces, want %d", len(traces), producerCount+3))
		}
		for _, trace := range traces {
			if len(trace) != 1 {
				panic(fmt.Sprintf("got %d spans in publish trace, want 1", len(trace)))
			}
			verifier.VerifyMQPublishAttributes(trace[0], "", "", "", "publish", topicName, "kafka")
		}
	}, producerCount+3)
}
