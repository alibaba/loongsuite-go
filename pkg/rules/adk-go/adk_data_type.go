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

package adkgo

import (
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
)

type adkGoInnerEnabler struct {
	enabled bool
}

func (e adkGoInnerEnabler) Enable() bool {
	return e.enabled
}

var adkGoEnabler = adkGoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ADK_GO_ENABLED") != "false"}

const (
	OperationChat        = "chat"
	OperationInvokeAgent = "invoke_agent"
	SystemGoogleADK      = "google_adk"
)

// adkLLMRequest holds the instrumentation data extracted from an ADK-Go LLM request.
type adkLLMRequest struct {
	operationName string
	modelName     string
	temperature   float64
	topP          float64
	topK          float64
	maxTokens     int64
	stopSequences []string
	isStream      bool
	inputTokens   int64
	inputMessages string
}

// adkLLMResponse holds the instrumentation data extracted from an ADK-Go LLM response.
type adkLLMResponse struct {
	responseModel     string
	usageInputTokens  int64
	usageOutputTokens int64
	finishReasons     []string
	outputMessages    string
}

// adkAgentRequest holds the instrumentation data for agent-level spans.
type adkAgentRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
	input         map[string]any
	output        map[string]any
}
