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

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func uintPtr(v uint) *uint {
	return &v
}

// This test verifies that the NewAPI relay integration compiles correctly
// with the loongsuite-go instrumentation. NewAPI's TextHelper operates
// within a gin HTTP handler context with complex internal state, so this
// test focuses on compilation verification and basic type usage.

func main() {
	gin.SetMode(gin.TestMode)

	// Create a test gin engine and context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a mock request
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Request = req

	// Set channel name in context (used by instrumentation)
	c.Set("channel_name", "test-provider")

	// Verify RelayInfo structure can be constructed
	info := &relaycommon.RelayInfo{
		UserId:          1,
		OriginModelName: "gpt-4",
		IsStream:        false,
		RequestId:       "req-test-123",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gpt-4",
		},
	}

	// Verify GeneralOpenAIRequest structure
	textReq := &dto.GeneralOpenAIRequest{
		Model:     "gpt-4",
		MaxTokens: uintPtr(100),
		Messages: []dto.Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	// Associate request with relay info
	info.Request = textReq

	fmt.Printf("NewAPI relay integration test passed: model=%s, user_id=%d, context=%v\n",
		textReq.Model, info.UserId, c != nil)
}
