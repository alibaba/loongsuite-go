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
	"log"
	"testing"
)

type testRequest struct {
}

type testResponse struct {
}

type netAttrsGetter struct {
}

func (n netAttrsGetter) GetUrlScheme(request testRequest) string {
	return "test"
}

func (n netAttrsGetter) GetUrlPath(request testRequest) string {
	return "test"
}

func (n netAttrsGetter) GetUrlQuery(request testRequest) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkType(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkTransport(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkProtocolName(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkProtocolVersion(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkLocalInetAddress(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkLocalPort(request testRequest, response testResponse) int {
	return 8080
}

func (n netAttrsGetter) GetNetworkPeerInetAddress(request testRequest, response testResponse) string {
	return "test"
}

func (n netAttrsGetter) GetNetworkPeerPort(request testRequest, response testResponse) int {
	return 8080
}

func TestUrlAttrsExtractorStart(t *testing.T) {
	urlExtractor := UrlAttrsExtractor[testRequest, testResponse, netAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = urlExtractor.OnStart(attrs, parentContext, testRequest{})
	if attrs[0].Key != semconv.URLSchemeKey || attrs[0].Value.AsString() != "test" {
		t.Fatalf("url scheme key should be test")
	}
	if attrs[1].Key != semconv.URLPathKey || attrs[1].Value.AsString() != "test" {
		t.Fatalf("url path should be test")
	}
	if attrs[2].Key != semconv.URLQueryKey || attrs[2].Value.AsString() != "test" {
		t.Fatalf("url query name should be test")
	}
}

func TestUrlAttrsExtractorEnd(t *testing.T) {
	urlExtractor := UrlAttrsExtractor[testRequest, testResponse, netAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = urlExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if len(attrs) != 0 {
		log.Fatal("attrs should be empty")
	}
}

func TestNetClientExtractorStart(t *testing.T) {
	netExtractor := NetworkAttrsExtractor[testRequest, testResponse, netAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = netExtractor.OnStart(attrs, parentContext, testRequest{})
	if len(attrs) != 0 {
		log.Fatal("attrs should be empty")
	}
}

func TestNetClientExtractorEnd(t *testing.T) {
	netExtractor := NetworkAttrsExtractor[testRequest, testResponse, netAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = netExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if attrs[0].Key != semconv.NetworkTransportKey || attrs[0].Value.AsString() != "test" {
		t.Fatalf("network transport key should be test")
	}
	if attrs[1].Key != semconv.NetworkTypeKey || attrs[1].Value.AsString() != "test" {
		t.Fatalf("network type should be test")
	}
	if attrs[2].Key != semconv.NetworkProtocolNameKey || attrs[2].Value.AsString() != "test" {
		t.Fatalf("network protocol name should be test")
	}
	if attrs[3].Key != semconv.NetworkProtocolVersionKey || attrs[3].Value.AsString() != "test" {
		t.Fatalf("network protocol version should be test")
	}
	if attrs[4].Key != semconv.NetworkLocalAddressKey || attrs[4].Value.AsString() != "test" {
		t.Fatalf("network local address should be test")
	}
	if attrs[5].Key != semconv.NetworkPeerAddressKey || attrs[5].Value.AsString() != "test" {
		t.Fatalf("network peer address should be test")
	}
	if attrs[6].Key != semconv.NetworkLocalPortKey || attrs[6].Value.AsInt64() != 8080 {
		t.Fatalf("network local port should be empty")
	}
	if attrs[7].Key != semconv.NetworkPeerPortKey || attrs[7].Value.AsInt64() != 8080 {
		t.Fatalf("network peer port should be empty")
	}
}
