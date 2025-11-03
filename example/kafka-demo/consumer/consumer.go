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

package consumer

import (
	"context"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"net/http"
	"strings"
	"time"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func (c *KafkaConsumer) Init(topic, group, endpoint string) {
	brokers := strings.Split(endpoint, ",")
	readerConfig := kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     group,
		Topic:       topic,
		StartOffset: kafka.LastOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		Dialer:      getDialer(),
	}

	c.reader = kafka.NewReader(readerConfig)
}

func (c *KafkaConsumer) ReceiveMessage(ctx context.Context) (kafka.Message, error) {
	message, err := c.reader.ReadMessage(ctx)
	return message, err
}

func (c *KafkaConsumer) ConsumerMsg(ctx context.Context, message kafka.Message) error {
	url := "http://www.baidu.com"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println(err.Error())
		time.Sleep(time.Second * 1)
		return err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		time.Sleep(time.Second * 1)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *KafkaConsumer) Close() {
	c.reader.Close()
}
