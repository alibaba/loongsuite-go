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
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel/trace"
)

//go:linkname publishToClientOnEnter github.com/mochi-mqtt/server/v2.publishToClientOnEnter
func publishToClientOnEnter(call api.CallContext, _ interface{}, ctx context.Context, cl interface{}, pk packets.Packet) {
	if !mqttEnabler.Enable() {
		return
	}

	// Extract client information using reflection
	// Client type from mochi-mqtt has ID field and Net.Remote field
	clientID := ""
	remote := ""

	if cl != nil {
		v := reflect.ValueOf(cl)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// Extract ID field
		if idField := v.FieldByName("ID"); idField.IsValid() && idField.Kind() == reflect.String {
			clientID = idField.String()
		}

		// Extract Net.Remote field
		if netField := v.FieldByName("Net"); netField.IsValid() {
			if remoteField := netField.FieldByName("Remote"); remoteField.IsValid() && remoteField.Kind() == reflect.String {
				remote = remoteField.String()
			}
		}
	}

	req := DeliverRequest{
		Packet:   pk,
		ClientID: clientID,
		Remote:   remote,
	}

	newCtx := StartDeliver(ctx, req)

	// Store context and request for exit hook
	call.SetData(map[string]interface{}{
		"ctx": newCtx,
		"req": req,
	})
}

//go:linkname publishToClientOnExit github.com/mochi-mqtt/server/v2.publishToClientOnExit
func publishToClientOnExit(call api.CallContext, err error) {
	if !mqttEnabler.Enable() {
		return
	}

	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}

	req, ok := data["req"].(DeliverRequest)
	if !ok {
		return
	}

	res := DeliverResponse{}

	EndDeliver(ctx, req, res, err)
}

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
