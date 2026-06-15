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

package message

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"testing"
)

type messageAttrsGetter struct {
}

func (m messageAttrsGetter) GetSystem(request testRequest) string {
	return "system"
}

func (m messageAttrsGetter) GetDestination(request testRequest) string {
	return "destination"
}

func (m messageAttrsGetter) GetDestinationTemplate(request testRequest) string {
	return "destination-template"
}

func (m messageAttrsGetter) IsTemporaryDestination(request testRequest) bool {
	return request.IsTemporaryDestination
}

func (m messageAttrsGetter) IsAnonymousDestination(request testRequest) bool {
	return request.IsAnonymousDestination
}

func (m messageAttrsGetter) GetConversationId(request testRequest) string {
	return "conversation-id"
}

func (m messageAttrsGetter) GetMessageBodySize(request testRequest) int64 {
	return 2024
}

func (m messageAttrsGetter) GetMessageEnvelopSize(request testRequest) int64 {
	return 2024
}

func (m messageAttrsGetter) GetMessageId(request testRequest, response testResponse) string {
	return "message-id"
}

func (m messageAttrsGetter) GetClientId(request testRequest) string {
	return "client-id"
}

func (m messageAttrsGetter) GetBatchMessageCount(request testRequest, response testResponse) int64 {
	return 2024
}

func (m messageAttrsGetter) GetMessageHeader(request testRequest, name string) []string {
	return []string{"header1", "header2"}
}

func (m messageAttrsGetter) GetDestinationPartitionId(request testRequest) string {
	return "partition-id"
}

func TestMessageGetSpanKey(t *testing.T) {
	messageExtractor := &MessageAttrsExtractor[testRequest, testResponse, messageAttrsGetter]{Operation: PUBLISH}
	if messageExtractor.GetSpanKey() != utils.PRODUCER_KEY {
		t.Fatalf("Should have returned producer key")
	}
	messageExtractor.Operation = RECEIVE
	if messageExtractor.GetSpanKey() != utils.CONSUMER_RECEIVE_KEY {
		t.Fatalf("Should have returned consumer receive key")
	}
	messageExtractor.Operation = PROCESS
	if messageExtractor.GetSpanKey() != utils.CONSUMER_PROCESS_KEY {
		t.Fatalf("Should have returned consumer process key")
	}
}

func TestMessageClientExtractorStartWithTemporaryDestination(t *testing.T) {
	messageExtractor := MessageAttrsExtractor[testRequest, testResponse, messageAttrsGetter]{Operation: PUBLISH}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = messageExtractor.OnStart(attrs, parentContext, testRequest{IsTemporaryDestination: true, IsAnonymousDestination: true})
	if attrs[0].Key != semconv.MessagingDestinationTemporaryKey || attrs[0].Value.AsBool() != true {
		t.Fatalf("temporary should be true")
	}
	if attrs[1].Key != semconv.MessagingDestinationNameKey || attrs[1].Value.AsString() != "(temporary)" {
		t.Fatalf("destination name should be temporary")
	}
	if attrs[2].Key != semconv.MessagingDestinationPartitionIDKey || attrs[2].Value.AsString() != "partition-id" {
		t.Fatalf("destination anonymous should be true")
	}
	if attrs[3].Key != semconv.MessagingDestinationAnonymousKey || attrs[3].Value.AsBool() != true {
		t.Fatalf("partition ID should be partition-id")
	}
	if attrs[4].Key != semconv.MessagingMessageConversationIDKey || attrs[4].Value.AsString() != "conversation-id" {
		t.Fatalf("conversation should be conversation-id")
	}
	if attrs[5].Key != semconv.MessagingMessageBodySizeKey || attrs[5].Value.AsInt64() != 2024 {
		t.Fatalf("message body size should be 2024")
	}
	if attrs[6].Key != semconv.MessagingMessageEnvelopeSizeKey || attrs[6].Value.AsInt64() != 2024 {
		t.Fatalf("message envelope size should be 2024")
	}
	if attrs[7].Key != semconv.MessagingClientIDKey || attrs[7].Value.AsString() != "client-id" {
		t.Fatalf("message client id should be client-id")
	}
	if attrs[8].Key != semconv.MessagingOperationNameKey || attrs[8].Value.AsString() != "publish" {
		t.Fatalf("message operation should be publish")
	}
	if attrs[9].Key != semconv.MessagingSystemKey || attrs[9].Value.AsString() != "system" {
		t.Fatalf("message system should be system")
	}
}

func TestMessageClientExtractorStartWithoutTemporaryDestination(t *testing.T) {
	messageExtractor := MessageAttrsExtractor[testRequest, testResponse, messageAttrsGetter]{Operation: PUBLISH}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = messageExtractor.OnStart(attrs, parentContext, testRequest{IsTemporaryDestination: false, IsAnonymousDestination: true})
	if attrs[0].Key != semconv.MessagingDestinationNameKey || attrs[0].Value.AsString() != "destination" {
		t.Fatalf("destination name should be destination")
	}
	if attrs[1].Key != semconv.MessagingDestinationTemplateKey || attrs[1].Value.AsString() != "destination-template" {
		t.Fatalf("destination template should be destination-template")
	}
	if attrs[2].Key != semconv.MessagingDestinationPartitionIDKey || attrs[2].Value.AsString() != "partition-id" {
		t.Fatalf("destination anonymous should be true")
	}
	if attrs[3].Key != semconv.MessagingDestinationAnonymousKey || attrs[3].Value.AsBool() != true {
		t.Fatalf("partition ID should be partition-id")
	}
	if attrs[4].Key != semconv.MessagingMessageConversationIDKey || attrs[4].Value.AsString() != "conversation-id" {
		t.Fatalf("conversation should be conversation-id")
	}
	if attrs[5].Key != semconv.MessagingMessageBodySizeKey || attrs[5].Value.AsInt64() != 2024 {
		t.Fatalf("message body size should be 2024")
	}
	if attrs[6].Key != semconv.MessagingMessageEnvelopeSizeKey || attrs[6].Value.AsInt64() != 2024 {
		t.Fatalf("message envelope size should be 2024")
	}
	if attrs[7].Key != semconv.MessagingClientIDKey || attrs[7].Value.AsString() != "client-id" {
		t.Fatalf("message client id should be client-id")
	}
	if attrs[8].Key != semconv.MessagingOperationNameKey || attrs[8].Value.AsString() != "publish" {
		t.Fatalf("message operation should be publish")
	}
	if attrs[9].Key != semconv.MessagingSystemKey || attrs[9].Value.AsString() != "system" {
		t.Fatalf("message system should be system")
	}
}

func TestMessageClientExtractorEnd(t *testing.T) {
	messageExtractor := MessageAttrsExtractor[testRequest, testResponse, messageAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = messageExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if attrs[0].Key != semconv.MessagingMessageIDKey || attrs[0].Value.AsString() != "message-id" {
		t.Fatalf("message id should be message-id")
	}
	if attrs[1].Key != semconv.MessagingBatchMessageCountKey || attrs[1].Value.AsInt64() != 2024 {
		t.Fatalf("messaging batch message count should be 2024")
	}
}
