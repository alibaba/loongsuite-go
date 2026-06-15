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
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"github.com/mochi-mqtt/server/v2/packets"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"
)

var (
	mqttEnabler = mqttInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_MQTT_ENABLED") != "false"}
	publishInst = newPublishInstrumenter()
	deliverInst = newDeliverInstrumenter()
)

type mqttInnerEnabler struct {
	enabled bool
}

func (g mqttInnerEnabler) Enable() bool {
	return g.enabled
}

func userPropGet(props []packets.UserProperty, key string) string {
	for i := len(props) - 1; i >= 0; i-- {
		if props[i].Key == key {
			return props[i].Val
		}
	}
	return ""
}

func userPropSet(props []packets.UserProperty, key, val string) []packets.UserProperty {
	return append(props, packets.UserProperty{Key: key, Val: val})
}

func userPropKeys(props []packets.UserProperty) []string {
	seen := make(map[string]struct{}, len(props))
	keys := make([]string, 0, len(props))
	for _, p := range props {
		if _, ok := seen[p.Key]; !ok {
			seen[p.Key] = struct{}{}
			keys = append(keys, p.Key)
		}
	}
	return keys
}

type ProducerCarrier struct {
	pk *packets.Packet
}

func (c ProducerCarrier) Get(key string) string {
	return userPropGet(c.pk.Properties.User, key)
}
func (c ProducerCarrier) Set(key, val string) {
	c.pk.Properties.User = userPropSet(c.pk.Properties.User, key, val)
}
func (c ProducerCarrier) Keys() []string {
	return userPropKeys(c.pk.Properties.User)
}

type ConsumerCarrier struct {
	pk *packets.Packet
}

func (c ConsumerCarrier) Get(key string) string {
	return userPropGet(c.pk.Properties.User, key)
}
func (c ConsumerCarrier) Set(key, val string) {
	c.pk.Properties.User = userPropSet(c.pk.Properties.User, key, val)
}
func (c ConsumerCarrier) Keys() []string {
	return userPropKeys(c.pk.Properties.User)
}

type PublishStatusExtractor struct{}

func (e *PublishStatusExtractor) Extract(span trace.Span, _ PublishRequest, _ PublishResponse, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type DeliverStatusExtractor struct{}

func (e *DeliverStatusExtractor) Extract(span trace.Span, _ DeliverRequest, _ DeliverResponse, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

type PublishAttrsGetter struct{}

func (g PublishAttrsGetter) GetSystem(PublishRequest) string {
	return "mqtt"
}
func (g PublishAttrsGetter) GetDestination(req PublishRequest) string {
	return req.Packet.TopicName
}
func (g PublishAttrsGetter) GetDestinationTemplate(PublishRequest) string {
	return ""
}
func (g PublishAttrsGetter) IsTemporaryDestination(PublishRequest) bool {
	return false
}
func (g PublishAttrsGetter) IsAnonymousDestination(PublishRequest) bool {
	return false
}
func (g PublishAttrsGetter) GetConversationId(PublishRequest) string {
	return ""
}
func (g PublishAttrsGetter) GetMessageBodySize(req PublishRequest) int64 {
	return int64(len(req.Packet.Payload))
}
func (g PublishAttrsGetter) GetMessageEnvelopSize(PublishRequest) int64 {
	return 0
}
func (g PublishAttrsGetter) GetMessageId(req PublishRequest, _ PublishResponse) string {
	if req.Packet.PacketID != 0 {
		return req.Packet.FormatID()
	}
	return ""
}
func (g PublishAttrsGetter) GetClientId(req PublishRequest) string {
	return req.ClientID
}
func (g PublishAttrsGetter) GetBatchMessageCount(PublishRequest, PublishResponse) int64 {
	return 1
}
func (g PublishAttrsGetter) GetMessageHeader(req PublishRequest, name string) []string {
	if v := userPropGet(req.Packet.Properties.User, name); v != "" {
		return []string{v}
	}
	return nil
}
func (g PublishAttrsGetter) GetDestinationPartitionId(PublishRequest) string { return "" }

type DeliverAttrsGetter struct{}

func (g DeliverAttrsGetter) GetSystem(DeliverRequest) string {
	return "mqtt"
}
func (g DeliverAttrsGetter) GetDestination(req DeliverRequest) string {
	return req.Packet.TopicName
}
func (g DeliverAttrsGetter) GetDestinationTemplate(DeliverRequest) string {
	return ""
}
func (g DeliverAttrsGetter) IsTemporaryDestination(DeliverRequest) bool {
	return false
}
func (g DeliverAttrsGetter) IsAnonymousDestination(DeliverRequest) bool {
	return false
}
func (g DeliverAttrsGetter) GetConversationId(DeliverRequest) string {
	return ""
}
func (g DeliverAttrsGetter) GetMessageBodySize(req DeliverRequest) int64 {
	return int64(len(req.Packet.Payload))
}
func (g DeliverAttrsGetter) GetMessageEnvelopSize(DeliverRequest) int64 {
	return 0
}
func (g DeliverAttrsGetter) GetMessageId(req DeliverRequest, _ DeliverResponse) string {
	if req.Packet.PacketID != 0 {
		return req.Packet.FormatID()
	}
	return ""
}
func (g DeliverAttrsGetter) GetClientId(req DeliverRequest) string {
	return req.ClientID
}
func (g DeliverAttrsGetter) GetBatchMessageCount(DeliverRequest, DeliverResponse) int64 {
	return 1
}
func (g DeliverAttrsGetter) GetMessageHeader(req DeliverRequest, name string) []string {
	if v := userPropGet(req.Packet.Properties.User, name); v != "" {
		return []string{v}
	}
	return nil
}
func (g DeliverAttrsGetter) GetDestinationPartitionId(DeliverRequest) string {
	return ""
}

type PublishAttrsExtractor struct{}

func (e *PublishAttrsExtractor) OnStart(attrs []attribute.KeyValue, ctx context.Context, req PublishRequest) ([]attribute.KeyValue, context.Context) {
	return append(attrs,
		attribute.String("messaging.system", "mqtt"),
		attribute.String("messaging.destination", req.Packet.TopicName),
		attribute.Int64("messaging.message.payload_size_bytes", int64(len(req.Packet.Payload))),
		attribute.Int("messaging.mqtt.qos", int(req.Packet.FixedHeader.Qos)),
		attribute.Bool("messaging.mqtt.retain", req.Packet.FixedHeader.Retain),
		attribute.Bool("messaging.mqtt.dup", req.Packet.FixedHeader.Dup),
		attribute.String("messaging.client_id", req.ClientID),
		attribute.String("messaging.client_remote", req.Remote),
	), ctx
}

func (e *PublishAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, _ PublishRequest, _ PublishResponse, err error) ([]attribute.KeyValue, context.Context) {
	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
	}
	return attrs, ctx
}

type DeliverAttrsExtractor struct{}

func (e *DeliverAttrsExtractor) OnStart(attrs []attribute.KeyValue, ctx context.Context, req DeliverRequest) ([]attribute.KeyValue, context.Context) {
	return append(attrs,
		attribute.String("messaging.system", "mqtt"),
		attribute.String("messaging.destination", req.Packet.TopicName),
		attribute.Int64("messaging.message.payload_size_bytes", int64(len(req.Packet.Payload))),
		attribute.Int("messaging.mqtt.qos", int(req.Packet.FixedHeader.Qos)),
		attribute.Bool("messaging.mqtt.retain", req.Packet.FixedHeader.Retain),
		attribute.Bool("messaging.mqtt.dup", req.Packet.FixedHeader.Dup),
		attribute.String("messaging.client_id", req.ClientID),
		attribute.String("messaging.client_remote", req.Remote),
	), ctx
}

func (e *DeliverAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, _ DeliverRequest, _ DeliverResponse, err error) ([]attribute.KeyValue, context.Context) {
	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
	}
	return attrs, ctx
}

func newPublishInstrumenter() instrumenter.Instrumenter[PublishRequest, PublishResponse] {
	builder := instrumenter.Builder[PublishRequest, PublishResponse]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MQTT_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[PublishRequest, PublishResponse]{
			Getter:        PublishAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[PublishRequest]{}).
		AddAttributesExtractor(&PublishAttrsExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[PublishRequest, PublishResponse, PublishAttrsGetter]{
			Operation: message.PUBLISH,
		}).
		SetSpanStatusExtractor(&PublishStatusExtractor{}).
		BuildPropagatingToDownstreamInstrumenter(
			func(req PublishRequest) propagation.TextMapCarrier {
				return ProducerCarrier{
					pk: req.Packet,
				}
			},
			otel.GetTextMapPropagator(),
		)
}

func newDeliverInstrumenter() instrumenter.Instrumenter[DeliverRequest, DeliverResponse] {
	builder := instrumenter.Builder[DeliverRequest, DeliverResponse]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MQTT_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[DeliverRequest, DeliverResponse]{
			Getter:        DeliverAttrsGetter{},
			OperationName: message.RECEIVE,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[DeliverRequest]{}).
		AddAttributesExtractor(&DeliverAttrsExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[DeliverRequest, DeliverResponse, DeliverAttrsGetter]{
			Operation: message.RECEIVE,
		}).
		SetSpanStatusExtractor(&DeliverStatusExtractor{}).
		BuildPropagatingFromUpstreamInstrumenter(
			func(req DeliverRequest) propagation.TextMapCarrier {
				return ConsumerCarrier{
					pk: &req.Packet,
				}
			},
			otel.GetTextMapPropagator(),
		)
}

func StartPublish(ctx context.Context, req PublishRequest) context.Context {
	if !mqttEnabler.Enable() {
		return ctx
	}
	return publishInst.Start(ctx, req)
}

func EndPublish(ctx context.Context, req PublishRequest, res PublishResponse, err error) {
	if !mqttEnabler.Enable() {
		return
	}
	publishInst.End(ctx, req, res, err)
}

func StartDeliver(ctx context.Context, req DeliverRequest) context.Context {
	if !mqttEnabler.Enable() {
		return ctx
	}
	return deliverInst.Start(ctx, req)
}

func EndDeliver(ctx context.Context, req DeliverRequest, res DeliverResponse, err error) {
	if !mqttEnabler.Enable() {
		return
	}
	deliverInst.End(ctx, req, res, err)
}
