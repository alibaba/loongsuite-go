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

package rpc

import "testing"

type spanTestRequest struct {
	System  string
	Service string
	Method  string
}

type spanTestGetter struct {
}

func (t spanTestGetter) GetSystem(request spanTestRequest) string {
	if request.System != "" {
		return request.System
	}
	return ""
}

func (t spanTestGetter) GetService(request spanTestRequest) string {
	if request.Service != "" {
		return request.Service
	}
	return ""
}

func (t spanTestGetter) GetMethod(request spanTestRequest) string {
	if request.Method != "" {
		return request.Method
	}
	return ""
}

func (t spanTestGetter) GetServerAddress(request spanTestRequest) string {
	return "test"
}

func TestExtractSpanName(t *testing.T) {
	r := RpcSpanNameExtractor[spanTestRequest]{Getter: spanTestGetter{}}
	spanName := r.Extract(spanTestRequest{Method: "method", Service: "service"})
	if spanName != "service/method" {
		t.Fatalf("extract span name extractor failed, expected 'service/method', got '%s'", spanName)
	}
	spanName = r.Extract(spanTestRequest{})
	if spanName != "RPC request" {
		t.Fatalf("extract span name extractor failed, expected 'RPC request', got '%s'", spanName)
	}
}
