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

package openai

import (
	"context"
	"encoding/json"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

// Hooks for github.com/openai/openai-go (official OpenAI SDK)

//go:linkname newChatCompletionOnEnter github.com/openai/openai-go.newChatCompletionOnEnter
func newChatCompletionOnEnter(call api.CallContext, r *openai.ChatCompletionService, ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) {
	if !openaiEnabler.Enable() {
		return
	}

	request := openaiRequest{
		operationName:    OperationNameChat,
		modelName:        body.Model,
		temperature:      body.Temperature.Value,
		topP:             body.TopP.Value,
		seed:             body.Seed.Value,
		maxTokens:        body.MaxTokens.Value,
		frequencyPenalty: body.FrequencyPenalty.Value,
		presencePenalty:  body.PresencePenalty.Value,
		uid:              body.User.Value,
		isStream:         false,
	}
	input, err := json.Marshal(body.Messages)
	if err == nil {
		request.inputMessages = string(input)
	}
	recorder := NewAIMetricsRecorder()
	instrumentedCtx := recorder.Start(ctx, request)

	// Store context and request in call data
	data := make(map[string]interface{})
	data["ctx"] = instrumentedCtx
	data["request"] = request
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname newChatCompletionOnExit github.com/openai/openai-go.newChatCompletionOnExit
func newChatCompletionOnExit(call api.CallContext, resp *openai.ChatCompletion, err error) {
	data := call.GetData().(map[string]interface{})
	if data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(openaiRequest)
	recorder, _ := data["recorder"].(*AIMetricsRecorder)

	if recorder == nil || ctx == nil {
		return
	}

	response := openaiResponse{}

	if err == nil && resp != nil {
		response.responseID = resp.ID
		response.responseModel = resp.Model
		response.usageTotalTokens = resp.Usage.TotalTokens
		response.usageInputTokens = resp.Usage.PromptTokens
		request.inputTokens = response.usageInputTokens
		response.usageOutputTokens = resp.Usage.CompletionTokens
		response.choiceCount = len(resp.Choices)
		var messages []openai.ChatCompletionMessage
		for _, choice := range resp.Choices {
			response.finishReasons = append(response.finishReasons, choice.FinishReason)
			messages = append(messages, choice.Message)
		}
		msgs, err1 := json.Marshal(messages)
		if err1 == nil {
			response.outputMessages = string(msgs)
		}
	}

	recorder.End(ctx, request, response, err)
}

//go:linkname officialNewChatCompletionStreamOnEnter github.com/openai/openai-go.officialNewChatCompletionStreamOnEnter
func officialNewChatCompletionStreamOnEnter(call api.CallContext, r *openai.ChatCompletionService, ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) {
	if !openaiEnabler.Enable() {
		return
	}
	request := openaiRequest{
		operationName:    OperationNameChat,
		modelName:        body.Model,
		temperature:      body.Temperature.Value,
		topP:             body.TopP.Value,
		seed:             body.Seed.Value,
		maxTokens:        body.MaxTokens.Value,
		frequencyPenalty: body.FrequencyPenalty.Value,
		presencePenalty:  body.PresencePenalty.Value,
		uid:              body.User.Value,
		isStream:         true,
	}
	input, err := json.Marshal(body.Messages)
	if err == nil {
		request.inputMessages = string(input)
	}
	recorder := NewAIMetricsRecorder()
	instrumentedCtx := recorder.Start(ctx, request)

	data := make(map[string]interface{})
	data["ctx"] = instrumentedCtx
	data["request"] = request
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(0, instrumentedCtx)
}

//go:linkname officialNewChatCompletionStreamOnExit github.com/openai/openai-go.officialNewChatCompletionStreamOnExit
func officialNewChatCompletionStreamOnExit(call api.CallContext, stream interface{}) {
	data := call.GetData().(map[string]interface{})
	if data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(openaiRequest)
	recorder, _ := data["recorder"].(*AIMetricsRecorder)

	if recorder == nil || ctx == nil {
		return
	}
	// For streaming, we record the start but the actual response data
	// will be collected as the stream is consumed
	response := openaiResponse{}
	recorder.End(ctx, request, response, nil)
}
