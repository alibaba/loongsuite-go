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

package go_openai

import "os"

// openaiInnerEnabler controls whether OpenAI monitoring is enabled
type openaiInnerEnabler struct {
	enabled bool
}

func (o openaiInnerEnabler) Enable() bool {
	return o.enabled
}

var openaiEnabler = openaiInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_OPENAI_ENABLED") != "false"}

const (
	OperationNameChat   = "chat"
	OperationNameStream = "chat.stream"
)

// openaiRequest represents the monitoring data for an OpenAI API request
type openaiRequest struct {
	uid              string
	operationName    string
	modelName        string
	frequencyPenalty float64
	presencePenalty  float64
	maxTokens        int64
	temperature      float64
	topP             float64
	seed             int64
	inputMessages    string
	isStream         bool
	inputTokens      int64
	stopSequences    []string
	serverAddress    string
}

// openaiResponse represents the monitoring data for an OpenAI API response
type openaiResponse struct {
	responseModel     string
	usageInputTokens  int64
	usageOutputTokens int64
	usageTotalTokens  int64
	responseID        string
	outputMessages    string
	choiceCount       int
	finishReasons     []string
}
