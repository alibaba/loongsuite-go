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
