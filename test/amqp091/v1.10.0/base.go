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
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
)

const queueName = "test_queue"
const exchange = "test_exchange"
const routingKey = "test_routing"

func initMQ() *amqp.Channel {

	conn, err := amqp.Dial("amqp://127.0.0.1:" + os.Getenv("RabbitMQ_PORT") + "/")
	if err != nil {
		panic(err)
	}

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	if err = channel.ExchangeDeclare(exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		panic(err)
	}
	if err := channel.QueueBind(queueName, routingKey, exchange, false, nil); err != nil {
		panic(err)
	}
	return channel
}
