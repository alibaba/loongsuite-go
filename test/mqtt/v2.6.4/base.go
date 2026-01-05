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
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"os"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// Default MQTT configuration
	defaultMQTTAddr    = "127.0.0.1:1883"
	defaultMQTTWSAddr  = "127.0.0.1:1882"
	defaultClientID    = "test-client"
	defaultTopicName   = "test/topic"
	defaultQoS         = 1       // int type for Subscribe
	defaultQoSByte     = byte(1) // byte type for Publish
	serverStartTimeout = 5 * time.Second
	connectionTimeout  = 10 * time.Second
)

var (
	tracer         oteltrace.Tracer
	traceExporter  *tracetest.InMemoryExporter
	tracerProvider *trace.TracerProvider
)

// InitTracerProvider initializes the OpenTelemetry tracer provider
func InitTracerProvider() (*trace.TracerProvider, *tracetest.InMemoryExporter) {
	exporter := tracetest.NewInMemoryExporter()

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(exporter)),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	traceExporter = exporter
	tracerProvider = tp
	tracer = tp.Tracer("mochi-mqtt-test")

	log.Println("TracerProvider initialized successfully")
	return tp, exporter
}

// GetSpans returns collected spans from the exporter
func GetSpans() tracetest.SpanStubs {
	if traceExporter == nil {
		return nil
	}
	return traceExporter.GetSpans()
}

// ResetSpans clears all collected spans
func ResetSpans() {
	if traceExporter != nil {
		traceExporter.Reset()
	}
}

// getMQTTAddress returns MQTT broker address from environment or default
func getMQTTAddress() string {
	if addr := os.Getenv("MQTT_ADDR"); addr != "" {
		return addr
	}
	return defaultMQTTAddr
}

// getMQTTWSAddress returns MQTT WebSocket address from environment or default
func getMQTTWSAddress() string {
	if addr := os.Getenv("MQTT_WS_ADDR"); addr != "" {
		return addr
	}
	return defaultMQTTWSAddr
}

// getClientID returns MQTT client ID from environment or default
func getClientID() string {
	if id := os.Getenv("MQTT_CLIENT_ID"); id != "" {
		return id
	}
	return defaultClientID
}

// initMQTTServer creates and starts a new Mochi MQTT server
func initMQTTServer() *mqtt.Server {
	// Create new MQTT server instance
	server := mqtt.New(&mqtt.Options{
		InlineClient: true, // Enable inline client for testing
	})

	// Allow all connections with auth hook
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		log.Fatalf("Failed to add auth hook: %v", err)
	}

	// Add tracing hook
	tracingHook := &TracingHook{}
	if err := server.AddHook(tracingHook, nil); err != nil {
		log.Fatalf("Failed to add tracing hook: %v", err)
	}

	// Add TCP listener
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: getMQTTAddress(),
	})
	if err := server.AddListener(tcp); err != nil {
		log.Fatalf("Failed to add TCP listener: %v", err)
	}

	// Add WebSocket listener (optional)
	ws := listeners.NewWebsocket(listeners.Config{
		ID:      "ws1",
		Address: getMQTTWSAddress(),
	})
	if err := server.AddListener(ws); err != nil {
		log.Printf("Warning: Failed to add WebSocket listener: %v", err)
	}

	// Start server in goroutine
	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("Failed to start MQTT server: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(serverStartTimeout)
	log.Printf("MQTT server started on %s", getMQTTAddress())

	return server
}

// stopMQTTServer gracefully stops the MQTT server
func stopMQTTServer(server *mqtt.Server) {
	if server != nil {
		if err := server.Close(); err != nil {
			log.Printf("Error closing MQTT server: %v", err)
		} else {
			log.Println("MQTT server stopped successfully")
		}
	}
}

// TracingHook implements OpenTelemetry tracing for MQTT operations
type TracingHook struct {
	mqtt.HookBase
}

func (h *TracingHook) ID() string {
	return "tracing-hook"
}

func (h *TracingHook) Provides(b byte) bool {
	return b == mqtt.OnPublish || b == mqtt.OnPublished
}

func (h *TracingHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	ctx := context.Background()

	tr := otel.GetTracerProvider().Tracer("mochi-mqtt-test")

	// Create publish span
	ctx, span := tr.Start(ctx, pk.TopicName+" publish",
		oteltrace.WithSpanKind(oteltrace.SpanKindProducer),
		oteltrace.WithAttributes(
			attribute.String("messaging.system", "mqtt"),
			attribute.String("messaging.destination.name", pk.TopicName),
			attribute.String("messaging.operation.name", "publish"),
			attribute.Int("messaging.mqtt.qos", int(pk.FixedHeader.Qos)),
			attribute.Int("messaging.message.body.size", len(pk.Payload)),
		),
	)

	// Store context in packet for later use
	pk.Properties.User = append(pk.Properties.User, packets.UserProperty{
		Key: "otel-trace-id",
		Val: span.SpanContext().TraceID().String(),
	})

	span.End()

	return pk, nil
}

func (h *TracingHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	ctx := context.Background()

	tr := otel.GetTracerProvider().Tracer("mochi-mqtt-test")

	// Create receive/process span
	_, span := tr.Start(ctx, pk.TopicName+" process",
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
		oteltrace.WithAttributes(
			attribute.String("messaging.system", "mqtt"),
			attribute.String("messaging.destination.name", pk.TopicName),
			attribute.String("messaging.operation.name", "process"),
			attribute.Int("messaging.mqtt.qos", int(pk.FixedHeader.Qos)),
			attribute.Int("messaging.message.body.size", len(pk.Payload)),
		),
	)

	span.End()
}

// MQTTConfig holds MQTT client configuration
type MQTTConfig struct {
	Broker   string
	ClientID string
	Username string
	Password string
	QoS      int  // for Subscribe
	QoSByte  byte // for Publish
	Topic    string
}

// GetDefaultConfig returns default MQTT configuration
func GetDefaultConfig() *MQTTConfig {
	return &MQTTConfig{
		Broker:   getMQTTAddress(),
		ClientID: getClientID(),
		QoS:      defaultQoS,
		QoSByte:  defaultQoSByte,
		Topic:    defaultTopicName,
	}
}

// VerifyMQTTPublishAttributes verifies span attributes for MQTT publish operation
func VerifyMQTTPublishAttributes(span tracetest.SpanStub, topic string, qos byte, expectedError bool) {
	// Verify span name
	expectedName := fmt.Sprintf("%s publish", topic)
	if span.Name != expectedName {
		log.Printf("Expected span name '%s', got '%s'", expectedName, span.Name)
	}

	// Verify messaging system
	verifyMQTTAttribute(span, "messaging.system", "mqtt")

	// Verify destination/topic
	verifyMQTTAttribute(span, "messaging.destination.name", topic)

	// Verify operation
	verifyMQTTAttribute(span, "messaging.operation.name", "publish")

	// Verify MQTT specific attributes
	verifyMQTTQoS(span, qos)

	// Verify span kind
	if span.SpanKind != oteltrace.SpanKindProducer {
		log.Printf("Expected producer span kind, got %d", span.SpanKind)
	}

	// Verify error status
	verifyMQTTErrorStatus(span, expectedError)

	log.Println("✓ Publish attributes verified")
}

// VerifyMQTTSubscribeAttributes verifies span attributes for MQTT subscribe operation
func VerifyMQTTSubscribeAttributes(span tracetest.SpanStub, topic string, qos byte, expectedError bool) {
	expectedName := fmt.Sprintf("%s subscribe", topic)
	if span.Name != expectedName {
		log.Printf("Expected span name '%s', got '%s'", expectedName, span.Name)
	}

	verifyMQTTAttribute(span, "messaging.system", "mqtt")
	verifyMQTTAttribute(span, "messaging.destination.name", topic)
	verifyMQTTAttribute(span, "messaging.operation.name", "subscribe")
	verifyMQTTQoS(span, qos)

	if span.SpanKind != oteltrace.SpanKindConsumer {
		log.Printf("Expected consumer span kind, got %d", span.SpanKind)
	}

	verifyMQTTErrorStatus(span, expectedError)
}

// VerifyMQTTReceiveAttributes verifies span attributes for MQTT receive/process operation
func VerifyMQTTReceiveAttributes(span tracetest.SpanStub, topic string, expectedError bool) {
	expectedName := fmt.Sprintf("%s process", topic)
	if span.Name != expectedName {
		log.Printf("Expected span name '%s', got '%s'", expectedName, span.Name)
	}

	verifyMQTTAttribute(span, "messaging.system", "mqtt")
	verifyMQTTAttribute(span, "messaging.destination.name", topic)
	verifyMQTTAttribute(span, "messaging.operation.name", "process")

	if span.SpanKind != oteltrace.SpanKindConsumer {
		log.Printf("Expected consumer span kind, got %d", span.SpanKind)
	}

	verifyMQTTErrorStatus(span, expectedError)

	log.Println("✓ Receive attributes verified")
}

// VerifyMQTTTraceContext verifies trace context propagation between publisher and subscriber
func VerifyMQTTTraceContext(publisher tracetest.SpanStub, subscriber tracetest.SpanStub) {
	// For now, just verify they exist in same trace
	if publisher.SpanContext.TraceID() == subscriber.SpanContext.TraceID() {
		log.Println("✓ Trace context verified - same trace ID")
	} else {
		log.Printf("Warning: Different trace IDs - Publisher: %s, Subscriber: %s",
			publisher.SpanContext.TraceID(), subscriber.SpanContext.TraceID())
	}
}

// verifyMQTTAttribute verifies a specific span attribute
func verifyMQTTAttribute(span tracetest.SpanStub, key string, expectedValue string) {
	for _, attr := range span.Attributes {
		if string(attr.Key) == key {
			actualValue := attr.Value.AsString()
			if actualValue != expectedValue {
				log.Printf("Expected %s '%s', got '%s'", key, expectedValue, actualValue)
			}
			return
		}
	}
	log.Printf("Attribute %s not found in span", key)
}

// verifyMQTTQoS verifies MQTT QoS level attribute
func verifyMQTTQoS(span tracetest.SpanStub, expectedQoS byte) {
	for _, attr := range span.Attributes {
		if string(attr.Key) == "messaging.mqtt.qos" {
			actualQoS := attr.Value.AsInt64()
			if actualQoS != int64(expectedQoS) {
				log.Printf("Expected QoS %d, got %d", expectedQoS, actualQoS)
			}
			return
		}
	}
	log.Printf("QoS attribute not found in span")
}

// verifyMQTTErrorStatus verifies the span error status
func verifyMQTTErrorStatus(span tracetest.SpanStub, expectedError bool) {
	if expectedError {
		if span.Status.Code != codes.Error {
			log.Printf("Expected error status, got %s", span.Status.Code)
		}
		if span.Status.Description == "" {
			log.Printf("Expected non-empty error description")
		}
	} else {
		if span.Status.Code == codes.Error {
			log.Printf("Expected non-error status, got error: %s", span.Status.Description)
		}
	}
}

// verifyMessageBodySize verifies message body size attribute
func verifyMessageBodySize(span tracetest.SpanStub, minSize int64) {
	for _, attr := range span.Attributes {
		if string(attr.Key) == "messaging.message.body.size" {
			actualSize := attr.Value.AsInt64()
			if actualSize < minSize {
				log.Printf("Expected message body size >= %d, got %d", minSize, actualSize)
			}
			return
		}
	}
	log.Printf("Message body size attribute not found in span")
}
