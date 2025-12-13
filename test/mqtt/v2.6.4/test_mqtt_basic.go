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
	"log"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	testTopic          = "test/basic"
	testMessageContent = "Hello Mochi MQTT"
	testQoS            = 1       // int type for Subscribe
	testQoSByte        = byte(1) // byte type for Publish
	messageWaitTime    = 2 * time.Second
)

var (
	messageReceived = false
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
	server := initMQTTServer()
	defer stopMQTTServer(server)

	// Subscribe to topic with callback
	err := server.Subscribe(testTopic, testQoS, messageHandler)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	log.Printf("Subscribed to topic: %s", testTopic)

	// Publish message
	err = server.Publish(testTopic, []byte(testMessageContent), false, testQoSByte)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Printf("Message published to topic: %s", testTopic)

	// Wait for message processing
	time.Sleep(messageWaitTime)

	// Verify message was received
	if !messageReceived {
		log.Fatal("Message was not received")
	}

	// Verify OpenTelemetry traces
	verifyBasicTraces(exporter)

	log.Println("Basic test completed successfully")
}

// messageHandler handles incoming MQTT messages
func messageHandler(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	log.Printf("Received message on topic %s: %s", pk.TopicName, string(pk.Payload))
	messageReceived = true
}

// verifyBasicTraces verifies the OpenTelemetry traces
func verifyBasicTraces(exporter *tracetest.InMemoryExporter) {
	// Get collected spans
	spans := exporter.GetSpans()

	log.Printf("Total spans collected: %d", len(spans))

	if len(spans) < 2 {
		log.Fatalf("Insufficient spans collected for verification. Expected at least 2, got %d", len(spans))
	}

	// The first span should be publish, second should be process
	publishSpan := spans[0]
	receiveSpan := spans[1]

	// Verify publish and receive spans
	VerifyMQTTPublishAttributes(publishSpan, testTopic, testQoSByte, false)
	VerifyMQTTReceiveAttributes(receiveSpan, testTopic, false)

	// Verify trace context
	VerifyMQTTTraceContext(publishSpan, receiveSpan)

	log.Println("✓ Basic trace verification passed")
}
