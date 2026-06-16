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

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	channel := initMQ()
	if err := channel.Confirm(false); err != nil {
		panic(err)
	}
	ack := make(chan uint64)
	nack := make(chan uint64)
	channel.NotifyConfirm(ack, nack)

	err := channel.Publish(
		exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			Body:         []byte("aabbcc"),
			DeliveryMode: 2,
		},
	)
	if err != nil {
		panic(err)
	}
	select {
	case <-ack:
		fmt.Println(true)
	case <-nack:
		fmt.Println(false)
	}

	msgCh, err := channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	if msg, ok := <-msgCh; ok {
		_ = msg.Ack(true)
	}

	destination := exchange
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.Assert(len(stubs) >= 2, "expected at least 2 traces, got %d", len(stubs))
		verifier.VerifyMQPublishAttributes(stubs[0][0], exchange, routingKey, queueName, "publish", destination, "rabbitmq")
		verifier.VerifyMQConsumeAttributes(stubs[1][0], exchange, routingKey, queueName, "receive", destination, "rabbitmq")
	}, 2)
}
