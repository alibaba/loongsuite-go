package consumer

import (
	"github.com/segmentio/kafka-go"

	"time"
)

// add kafka dialer like use aliyun kafka
func getDialer() *kafka.Dialer {
	return &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		//TLS:           getTlsConfig(),
		//SASLMechanism: getSaslMechanism(),
	}
}
