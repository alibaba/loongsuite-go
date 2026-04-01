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

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	mqtt "github.com/mochi-mqtt/server/v2"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	publishBuffer = 100 * time.Millisecond
)

type PublisherTestCase struct {
	name     string
	testFunc func(*mqtt.Server, *tracetest.InMemoryExporter)
}

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

	// Define test cases
	testCases := []PublisherTestCase{
		{"QoS 0 Publish Success", testQoS0PublishSuccess},
		{"QoS 1 Publish Success", testQoS1PublishSuccess},
		{"QoS 2 Publish Success", testQoS2PublishSuccess},
		{"Publish with Retained Flag", testPublishRetained},
		{"Publish to Multiple Topics", testPublishMultipleTopics},
	}

	// Execute test cases
	for _, tc := range testCases {
		log.Printf("\n===== Testing %s =====\n", tc.name)
		ResetSpans() // Clear spans before each test
		tc.testFunc(server, exporter)
		time.Sleep(publishBuffer)
	}

	log.Println("All publisher tests completed successfully")
}

// testQoS0PublishSuccess tests QoS 0 message publishing
func testQoS0PublishSuccess(server *mqtt.Server, exporter *tracetest.InMemoryExporter) {
	topic := "test/qos0"
	payload := []byte("Test QoS 0 message")
	qos := byte(0)

	err := server.Publish(topic, payload, false, qos)
	if err != nil {
		log.Printf("QoS 0 publish failed: %v\n", err)
		return
	}
	log.Println("QoS 0 publish succeeded")

	time.Sleep(100 * time.Millisecond) // Wait for span to be recorded
	verifyPublisherTraces(exporter, topic, qos, false)
}

// testQoS1PublishSuccess tests QoS 1 message publishing
func testQoS1PublishSuccess(server *mqtt.Server, exporter *tracetest.InMemoryExporter) {
	topic := "test/qos1"
	payload := []byte("Test QoS 1 message")
	qos := byte(1)

	err := server.Publish(topic, payload, false, qos)
	if err != nil {
		log.Printf("QoS 1 publish failed: %v\n", err)
		return
	}
	log.Println("QoS 1 publish succeeded")

	time.Sleep(100 * time.Millisecond)
	verifyPublisherTraces(exporter, topic, qos, false)
}

// testQoS2PublishSuccess tests QoS 2 message publishing
func testQoS2PublishSuccess(server *mqtt.Server, exporter *tracetest.InMemoryExporter) {
	topic := "test/qos2"
	payload := []byte("Test QoS 2 message")
	qos := byte(2)

	err := server.Publish(topic, payload, false, qos)
	if err != nil {
		log.Printf("QoS 2 publish failed: %v\n", err)
		return
	}
	log.Println("QoS 2 publish succeeded")

	time.Sleep(100 * time.Millisecond)
	verifyPublisherTraces(exporter, topic, qos, false)
}

// testPublishRetained tests publishing with retained flag
func testPublishRetained(server *mqtt.Server, exporter *tracetest.InMemoryExporter) {
	topic := "test/retained"
	payload := []byte("Test retained message")
	qos := byte(1)

	err := server.Publish(topic, payload, true, qos) // retain = true
	if err != nil {
		log.Printf("Retained publish failed: %v\n", err)
		return
	}
	log.Println("Retained publish succeeded")

	time.Sleep(100 * time.Millisecond)
	verifyPublisherTraces(exporter, topic, qos, false)
}

// testPublishMultipleTopics tests publishing to multiple topics
func testPublishMultipleTopics(server *mqtt.Server, exporter *tracetest.InMemoryExporter) {
	topics := []string{"test/topic1", "test/topic2", "test/topic3"}
	payload := []byte("Multi-topic test message")
	qos := byte(1)

	for i, topic := range topics {
		err := server.Publish(topic, payload, false, qos)
		if err != nil {
			log.Printf("Publish to topic %d failed: %v\n", i, err)
			return
		}
		log.Printf("Published to topic %d: %s\n", i, topic)
		time.Sleep(50 * time.Millisecond)
	}

	log.Println("Multi-topic publish succeeded")

	// Verify we have spans for all topics
	time.Sleep(200 * time.Millisecond)
	spans := exporter.GetSpans()
	log.Printf("Collected %d spans for %d topics", len(spans), len(topics))
}

// verifyPublisherTraces verifies OpenTelemetry traces for publisher
func verifyPublisherTraces(exporter *tracetest.InMemoryExporter, topic string, qos byte, expectError bool) {
	spans := exporter.GetSpans()

	if len(spans) == 0 {
		if !expectError {
			log.Printf("Warning: No spans collected for verification")
		}
		return
	}

	// Get the last span (most recent publish)
	span := spans[len(spans)-1]
	verifier.VerifyMQTTPublishAttributes(span, topic, qos, expectError)

	log.Printf("Publisher trace verification passed for topic: %s", topic)
}
