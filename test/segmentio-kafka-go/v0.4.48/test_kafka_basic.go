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

package main

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()

	// Initialize producer with cleanup
	producer := initProducer()
	defer producer.Close()

	// Initialize consumer with cleanup
	consumer := initConsumer()
	defer consumer.Close()

	// Send message
	messages := []kafka.Message{kafka.Message{
		Value: []byte("hello world1"),
	}, kafka.Message{
		Value: []byte("hello world2"),
	}}
	if err := producer.WriteMessages(ctx, messages...); err != nil {
		panic(err)
	}

	// Read message
	_, err := consumer.ReadMessage(context.Background())
	if err != nil {
		panic(err)
	}
	_, err = consumer.ReadMessage(context.Background())
	if err != nil {
		panic(err)
	}

	// Verify OpenTelemetry traces
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyMQPublishAttributes(stubs[0][0], "", "", "", "publish", topicName, "kafka")
		verifier.VerifyMQConsumeAttributes(stubs[0][1], "", "", "", "process", topicName, "kafka")
		verifier.VerifyMQConsumeAttributes(stubs[0][2], "", "", "", "process", topicName, "kafka")
	}, 1)
}
