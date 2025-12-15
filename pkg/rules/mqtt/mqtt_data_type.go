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

import "github.com/mochi-mqtt/server/v2/packets"

// PublishRequest represents a request to publish a message from an MQTT client to the broker.
// It is used in MQTT instrumentation to capture the details of a publish operation initiated by a client.
// The Packet field contains the MQTT publish packet, ClientID identifies the publishing client,
// and Remote provides the sender's remote address.
type PublishRequest struct {
	Packet   *packets.Packet // The MQTT publish packet
	ClientID string          // ID of the client publishing the message
	Remote   string          // sender remote address
}

// PublishResponse represents the result of processing a PublishRequest on the producer (client) side.
// This type is used as a placeholder for future extensions where additional response information
// may be needed after a publish operation. Currently, it does not contain any fields.
type PublishResponse struct{}

// DeliverRequest represents a request to deliver a published message from the broker to a subscriber.
// It is used in MQTT instrumentation to capture the details of a message delivery operation.
// The Packet field contains the publish packet being delivered, ClientID identifies the subscriber client,
// and Remote provides the subscriber's remote address.
type DeliverRequest struct {
	Packet   packets.Packet // publish packet being delivered
	ClientID string         // subscriber client id
	Remote   string         // subscriber remote address
}

// DeliverResponse represents the result of processing a DeliverRequest on the consumer (subscriber) side.
// This type is used as a placeholder for future extensions where additional response information
// may be needed after a delivery operation. Currently, it does not contain any fields.
type DeliverResponse struct{}
