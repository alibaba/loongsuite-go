// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package net

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"strings"
)

// TODO: remove server.address and put it into NetworkAttributesExtractor

type NetworkAttrsExtractor[REQUEST any, RESPONSE any, GETTER NetworkAttrsGetter[REQUEST, RESPONSE]] struct {
	Getter GETTER
}

func (i *NetworkAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	return attributes, parentContext
}

func (i *NetworkAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.NetworkTransportKey,
		Value: attribute.StringValue(i.Getter.GetNetworkTransport(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.NetworkTypeKey,
		Value: attribute.StringValue(strings.ToLower(i.Getter.GetNetworkType(request, response))),
	}, attribute.KeyValue{
		Key:   semconv.NetworkProtocolNameKey,
		Value: attribute.StringValue(strings.ToLower(i.Getter.GetNetworkProtocolName(request, response))),
	}, attribute.KeyValue{
		Key:   semconv.NetworkProtocolVersionKey,
		Value: attribute.StringValue(strings.ToLower(i.Getter.GetNetworkProtocolVersion(request, response))),
	}, attribute.KeyValue{
		Key:   semconv.NetworkLocalAddressKey,
		Value: attribute.StringValue(i.Getter.GetNetworkLocalInetAddress(request, response)),
	}, attribute.KeyValue{
		Key:   semconv.NetworkPeerAddressKey,
		Value: attribute.StringValue(i.Getter.GetNetworkPeerInetAddress(request, response)),
	})
	localPort := i.Getter.GetNetworkLocalPort(request, response)
	if localPort > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.NetworkLocalPortKey,
			Value: attribute.IntValue(localPort),
		})
	}
	peerPort := i.Getter.GetNetworkPeerPort(request, response)
	if peerPort > 0 {
		attributes = append(attributes, attribute.KeyValue{
			Key:   semconv.NetworkPeerPortKey,
			Value: attribute.IntValue(peerPort),
		})
	}
	return attributes, context
}

type UrlAttrsExtractor[REQUEST any, RESPONSE any, GETTER UrlAttrsGetter[REQUEST]] struct {
	Getter GETTER
	// TODO: add scheme provider for extension
}

func (u *UrlAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes, attribute.KeyValue{
		Key:   semconv.URLSchemeKey,
		Value: attribute.StringValue(u.Getter.GetUrlScheme(request)),
	}, attribute.KeyValue{
		Key:   semconv.URLPathKey,
		Value: attribute.StringValue(u.Getter.GetUrlPath(request)),
	}, attribute.KeyValue{
		Key:   semconv.URLQueryKey,
		Value: attribute.StringValue(u.Getter.GetUrlQuery(request)),
	})
	return attributes, parentContext
}

func (u *UrlAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, context
}
