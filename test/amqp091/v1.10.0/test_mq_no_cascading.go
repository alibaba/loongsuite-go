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

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	channel := initMQ()
	var err error
	if err = channel.Confirm(false); err != nil {
		panic(err)
	}
	var ack = make(chan uint64)
	var nack = make(chan uint64)
	channel.NotifyConfirm(ack, nack)

	_, err = channel.PublishWithDeferredConfirm(exchange, routingKey, true, false,
		amqp091.Publishing{Body: []byte("aabbcc"), DeliveryMode: 2})
	if err != nil {
		panic(err)
	}
	select {
	case <-ack:
		fmt.Println(true)
	case <-nack:
		fmt.Println(false)
	}

	msgChanl, err := channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	if msg, ok := <-msgChanl; ok {
		msg.Ack(true)
	}

	destination := exchange + ":" + routingKey
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyMQPublishAttributes(stubs[0][0], exchange, routingKey, queueName, "publish", destination, "rabbitmq")
		verifier.VerifyMQConsumeAttributes(stubs[1][0], exchange, routingKey, queueName, "receive", destination, "rabbitmq")
	}, 2)
}
