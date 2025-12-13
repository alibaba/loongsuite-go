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

package mqtt

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

// Enabled returns whether MQTT instrumentation is enabled
// It checks the OTEL_INSTRUMENTATION_MQTT_ENABLED environment variable
func Enabled() bool { return mqttEnabler.Enable() }

// ProducerInstrumenter returns the instrumenter instance for MQTT producer (publish) operations
// The returned interface provides Start and End methods for creating and completing trace spans
// when a client publishes messages to the broker
func ProducerInstrumenter() interface {
	Start(context.Context, PublishRequest, ...trace.SpanStartOption) context.Context
	End(context.Context, PublishRequest, PublishResponse, error, ...trace.SpanEndOption)
} {
	return publishInst
}

// ConsumerInstrumenter returns the instrumenter instance for MQTT consumer (deliver) operations
// The returned interface provides Start and End methods for creating and completing trace spans
// when the broker delivers messages to subscribed clients
func ConsumerInstrumenter() interface {
	Start(context.Context, DeliverRequest, ...trace.SpanStartOption) context.Context
	End(context.Context, DeliverRequest, DeliverResponse, error, ...trace.SpanEndOption)
} {
	return deliverInst
}

// StartProducer starts a new trace span for an MQTT publish operation
// It should be called when a client publishes a message to the broker
func StartProducer(ctx context.Context, req PublishRequest) context.Context {
	return StartPublish(ctx, req)
}

// EndProducer completes the trace span for an MQTT publish operation
// It should be called when the publish operation finishes, either successfully or with an error
func EndProducer(ctx context.Context, req PublishRequest, res PublishResponse, err error) {
	EndPublish(ctx, req, res, err)
}

// StartConsumer starts a new trace span for an MQTT message delivery operation
// It should be called when the broker is about to deliver a message to a subscribed client
func StartConsumer(ctx context.Context, req DeliverRequest) context.Context {
	return StartDeliver(ctx, req)
}

// EndConsumer completes the trace span for an MQTT message delivery operation
// It should be called when the delivery operation finishes, either successfully or with an error
func EndConsumer(ctx context.Context, req DeliverRequest, res DeliverResponse, err error) {
	EndDeliver(ctx, req, res, err)
}
