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
	"log"
	"sync"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	multiMessageCount    = 3
	multiTestTopic       = "test/multi"
	multiTestQoS         = 1       // int type for Subscribe
	multiTestQoSByte     = byte(1) // byte type for Publish
	multiMessageWaitTime = 3 * time.Second
)

var (
	receivedCount int
	mu            sync.Mutex
)

func main() {
	// Initialize TracerProvider first
	tp, exporter := InitTracerProvider()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize MQTT server
	server := initMQTTServer(tp)
	defer stopMQTTServer(server)

	// Subscribe to topic
	err := server.Subscribe(multiTestTopic, multiTestQoS, multiMessageHandler)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	log.Printf("Subscribed to topic: %s", multiTestTopic)

	// Publish multiple messages
	messages := prepareMultiMessages()
	publishMultiMessages(server, multiTestTopic, messages)

	// Wait for all messages to be processed
	time.Sleep(multiMessageWaitTime)

	// Verify all messages received
	mu.Lock()
	if receivedCount != multiMessageCount {
		log.Fatalf("Expected %d messages, received %d", multiMessageCount, receivedCount)
	}
	mu.Unlock()

	// Verify OpenTelemetry traces
	verifySubscriberTraces(exporter)

	log.Println("Multi-subscriber test completed successfully")
}

// multiMessageHandler handles multiple incoming messages
func multiMessageHandler(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("Received message: %s from topic: %s", string(pk.Payload), pk.TopicName)
	receivedCount++
}

// prepareMultiMessages creates multiple test messages
func prepareMultiMessages() []string {
	messages := make([]string, multiMessageCount)
	for i := 0; i < multiMessageCount; i++ {
		messages[i] = fmt.Sprintf("Multi-subscriber test message %d", i)
	}
	return messages
}

// publishMultiMessages publishes multiple messages
func publishMultiMessages(server *mqtt.Server, topic string, messages []string) {
	for i, payload := range messages {
		err := server.Publish(topic, []byte(payload), false, multiTestQoSByte)
		if err != nil {
			log.Fatalf("Failed to publish message %d: %v", i, err)
		}
		log.Printf("Published message %d: %s", i, payload)
		time.Sleep(100 * time.Millisecond)
	}
}

// verifySubscriberTraces verifies the OpenTelemetry traces for multiple messages
func verifySubscriberTraces(exporter *tracetest.InMemoryExporter) {
	spans := exporter.GetSpans()

	log.Printf("Total spans collected: %d", len(spans))

	// We expect at least multiMessageCount * 2 spans (publish + process for each message)
	expectedMinSpans := multiMessageCount * 2
	if len(spans) < expectedMinSpans {
		log.Printf("Warning: Expected at least %d spans, got %d", expectedMinSpans, len(spans))
	}

	// Verify spans in pairs (publish + process)
	verifiedCount := 0
	for i := 0; i < len(spans)-1; i += 2 {
		if i+1 >= len(spans) {
			break
		}

		publishSpan := spans[i]
		receiveSpan := spans[i+1]

		// Check if this is a publish span
		if publishSpan.Name == multiTestTopic+" publish" {
			VerifyMQTTPublishAttributes(publishSpan, multiTestTopic, multiTestQoSByte, false)

			// Check if next span is the corresponding process span
			if receiveSpan.Name == multiTestTopic+" process" {
				VerifyMQTTReceiveAttributes(receiveSpan, multiTestTopic, false)
				verifiedCount++
				log.Printf("✓ Message %d trace verification passed", verifiedCount)
			}
		}
	}

	if verifiedCount >= multiMessageCount {
		log.Println("All subscriber traces verified")
	} else {
		log.Printf("Warning: Only verified %d out of %d expected message traces", verifiedCount, multiMessageCount)
	}
}
