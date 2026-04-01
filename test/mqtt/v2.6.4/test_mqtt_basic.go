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
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	testTopic          = "test/basic"
	testMessageContent = "Hello Mochi MQTT"
	testQoS            = 1       // int type for Subscribe
	testQoSByte        = byte(1) // byte type for Publish
	messageWaitTime    = 3 * time.Second
)

var (
	messageReceived = false
)

func main() {
	// Initialize TracerProvider first
	tp, exporter := InitTracerProvider()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Shutting down tracer provider...")
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	server := initMQTTServer(tp)
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

	time.Sleep(messageWaitTime)

	log.Println("Forcing flush of pending spans...")

	for i := 0; i < 3; i++ {
		log.Println("Forcing flush of pending spans...")
		flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := tp.ForceFlush(flushCtx); err != nil {
			log.Printf("Warning: Failed to force flush spans (attempt %d): %v", i+1, err)
		}
		flushCancel()
		time.Sleep(500 * time.Millisecond)
	}

	// Additional wait for export completion
	time.Sleep(1 * time.Second)

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
	traceID := ""
	for _, prop := range pk.Properties.User {
		if prop.Key == "otel-trace-id" {
			traceID = prop.Val
			break
		}
	}

	if traceID != "" {
		log.Printf(" trace_id=%sReceived message on topic %s: %s", traceID, pk.TopicName, string(pk.Payload))
	} else {
		log.Printf("Received message on topic %s: %s", pk.TopicName, string(pk.Payload))
	}

	messageReceived = true
}

// verifyBasicTraces verifies the OpenTelemetry traces
func verifyBasicTraces(exporter *tracetest.InMemoryExporter) {
	// Get collected spans
	spans := exporter.GetSpans()

	log.Printf("Total spans collected: %d", len(spans))

	for i, span := range spans {
		log.Printf("Span %d: Name=%s, Kind=%v, TraceID=%s",
			i, span.Name, span.SpanKind, span.SpanContext.TraceID())
	}

	if len(spans) < 2 {
		log.Fatalf("Insufficient spans collected for verification. Expected at least 2, got %d", len(spans))
	}

	// The first span should be publish, second should be process
	publishSpan := spans[0]
	receiveSpan := spans[1]

	// Verify publish and receive spans
	verifier.VerifyMQTTPublishAttributes(publishSpan, testTopic, testQoSByte, false)
	verifier.VerifyMQTTReceiveAttributes(receiveSpan, testTopic, false)

	// Verify trace context
	verifier.VerifyMQTTTraceContext(publishSpan, receiveSpan)

	log.Println("✓ Basic trace verification passed")
}
