package main

import (
	"github.com/segmentio/kafka-go"

	"time"
)

var (
	kafkaEndpoint = "localhost:9093"
	topicPrefix   = "kafka-go-agent-it"
	groupPrefix   = "kafka-go-agent-it-group"
)

func getDialer() *kafka.Dialer {
	return &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		//TLS:           getTlsConfig(),
		//SASLMechanism: getSaslMechanism(),
	}
}
