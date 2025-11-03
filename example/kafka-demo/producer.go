package main

import (
	"context"
	kafka "github.com/segmentio/kafka-go"
	"strings"
	"time"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

func (p *kafkaProducer) init(topic, endpoint string) {
	brokers := strings.Split(endpoint, ",")
	writerConfig := kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Dialer:       getDialer(),
		BatchTimeout: 20 * time.Millisecond,
	}

	p.writer = kafka.NewWriter(writerConfig)
}

func (p *kafkaProducer) sendMessage(ctx context.Context, msg kafka.Message) error {
	return p.writer.WriteMessages(ctx, msg)
}

func (p *kafkaProducer) close() {
	p.writer.Close()
}
