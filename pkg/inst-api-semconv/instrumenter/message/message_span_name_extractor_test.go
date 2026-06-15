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
	"testing"
)

type testRequest struct {
	IsTemporaryDestination bool
	IsAnonymousDestination bool
	Destination            string
}

type testResponse struct {
}

type testGetter struct {
}

func (t testGetter) GetSystem(request testRequest) string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetDestination(request testRequest) string {
	if request.Destination != "" {
		return request.Destination
	}
	return ""
}

func (t testGetter) GetDestinationTemplate(request testRequest) string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) IsTemporaryDestination(request testRequest) bool {
	return request.IsTemporaryDestination
}

func (t testGetter) IsAnonymousDestination(request testRequest) bool {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetConversationId(request testRequest) string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetMessageBodySize(request testRequest) int64 {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetMessageEnvelopSize(request testRequest) int64 {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetMessageId(request testRequest, response testResponse) string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetClientId(request testRequest) string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetBatchMessageCount(request testRequest, response testResponse) int64 {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetMessageHeader(request testRequest, name string) []string {
	//TODO implement me
	panic("implement me")
}

func (t testGetter) GetDestinationPartitionId(request testRequest) string {
	return "partition-id"
}

func TestExtractSpanName(t *testing.T) {
	r := MessageSpanNameExtractor[testRequest, testResponse]{Getter: testGetter{}}
	spanName := r.Extract(testRequest{IsTemporaryDestination: true, Destination: "Destination"})
	if spanName != "(temporary)" {
		t.Fatalf("extract span name failed: expected (temporary) but got %s", spanName)
	}
	spanName = r.Extract(testRequest{IsTemporaryDestination: false, Destination: ""})
	if spanName != "unknown" {
		t.Fatalf("extract span name failed: expected unknown but got %s", spanName)
	}
}

func TestExtractSpanNameWithOperationName(t *testing.T) {
	r := MessageSpanNameExtractor[testRequest, testResponse]{Getter: testGetter{}, OperationName: PUBLISH}
	spanName := r.Extract(testRequest{IsTemporaryDestination: true, Destination: "Destination"})
	if spanName != "(temporary) publish" {
		t.Fatalf("extract span name failed: expected (temporary) publish but got %s", spanName)
	}
}
