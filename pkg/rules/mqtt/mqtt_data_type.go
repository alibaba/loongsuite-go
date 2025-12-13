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

// PublishRequest represents an inbound publishing from client to broker
type PublishRequest struct {
	Packet   *packets.Packet // The MQTT publish packet
	ClientID string          // ID of the client publishing the message
	Remote   string          // sender remote address
}

// PublishResponse is a placeholder for producer-side result.
type PublishResponse struct{}

// DeliverRequest represents a broker -> subscriber delivery.
type DeliverRequest struct {
	Packet   packets.Packet // publish packet being delivered
	ClientID string         // subscriber client id
	Remote   string         // subscriber remote address
}

// DeliverResponse is a placeholder for consumer-side result.
type DeliverResponse struct{}
