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
	"context"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"kafka-demo/consumer"
	"time"
)

var milli = time.Now().UnixMilli()

func main() {
	index := int(milli % 10)
	topic := fmt.Sprintf("%s-%v", topicPrefix, index)
	group := fmt.Sprintf("%s-%v", groupPrefix, index)
	fmt.Printf("Topic is %s\n", topic)
	fmt.Printf("Group is %s\n", group)

	go SetMessage()
	consumer1 := consumer.KafkaConsumer{}
	consumer1.Init(topic, group, kafkaEndpoint)
	defer consumer1.Close()

	ctx := context.Background()
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:

			ctx = context.Background()
			msg, err := consumer1.ReceiveMessage(ctx)
			if err != nil {
				panic(fmt.Sprintf("Cannot receive message, cause %v.\n", err))
			}
			consumer1.ConsumerMsg(ctx, msg)
		}
	}
}

func SetMessage() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			index := int(milli % 10)
			topic := fmt.Sprintf("%s-%v", topicPrefix, index)
			group := fmt.Sprintf("%s-%v", groupPrefix, index)
			fmt.Printf("Topic is %s\n", topic)
			fmt.Printf("Group is %s\n", group)
			producer := kafkaProducer{}
			producer.init(topic, kafkaEndpoint)
			defer producer.close()
			key := fmt.Sprintf("%s-%v", "kafka-it", milli)
			err := producer.sendMessage(context.Background(), kafka.Message{
				Key:   []byte(key),
				Value: []byte(fmt.Sprint("foobar")),
			})
			if err != nil {
				panic(fmt.Sprintf("Cannot send message, cause %v.\n", err))
			}
			time.Sleep(1 * time.Second)
		}
	}

}
