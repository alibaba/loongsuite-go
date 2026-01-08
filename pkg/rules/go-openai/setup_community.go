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

import (
	"context"
	"encoding/json"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	openai "github.com/sashabaranov/go-openai"
	_ "unsafe"
)

// Hooks for github.com/sashabaranov/go-openai (community SDK)
// Note: This also covers forks like github.com/meguminnnnnnnnn/go-openai

//go:linkname communityCreateChatCompletionOnEnter github.com/sashabaranov/go-openai.communityCreateChatCompletionOnEnter
func communityCreateChatCompletionOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request openai.CompletionRequest) {
	if !openaiEnabler.Enable() {
		return
	}
	req := openaiRequest{
		operationName:    OperationNameChat,
		modelName:        request.Model,
		temperature:      float64(request.Temperature),
		topP:             float64(request.TopP),
		maxTokens:        int64(request.MaxTokens),
		frequencyPenalty: float64(request.FrequencyPenalty),
		presencePenalty:  float64(request.PresencePenalty),
		uid:              request.User,
		isStream:         request.Stream,
	}
	if request.Seed != nil {
		req.seed = int64(*request.Seed)
	}
	input, err := json.Marshal(request.Prompt)
	if err == nil {
		req.inputMessages = string(input)
	}

	recorder := NewAIMetricsRecorder()
	instrumentedCtx := recorder.Start(ctx, req)

	data := make(map[string]interface{})
	data["ctx"] = instrumentedCtx
	data["request"] = req
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname communityCreateChatCompletionOnExit github.com/sashabaranov/go-openai.communityCreateChatCompletionOnExit
func communityCreateChatCompletionOnExit(call api.CallContext, resp openai.CompletionResponse, err error) {
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

	if err == nil {
		response.responseID = resp.ID
		response.responseModel = resp.Model
		response.usageTotalTokens = int64(resp.Usage.TotalTokens)
		response.usageInputTokens = int64(resp.Usage.PromptTokens)
		request.inputTokens = response.usageInputTokens
		response.usageOutputTokens = int64(resp.Usage.CompletionTokens)
		response.choiceCount = len(resp.Choices)
		var msgs []string
		for _, choice := range resp.Choices {
			response.finishReasons = append(response.finishReasons, choice.FinishReason)
			msgs = append(msgs, choice.Text)
		}
	}

	recorder.End(ctx, request, response, err)
}

//go:linkname communityCreateChatCompletionStreamOnEnter github.com/sashabaranov/go-openai.communityCreateChatCompletionStreamOnEnter
func communityCreateChatCompletionStreamOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request openai.ChatCompletionRequest) {
	if !openaiEnabler.Enable() {
		return
	}

	req := openaiRequest{
		operationName:    OperationNameChat,
		modelName:        request.Model,
		temperature:      float64(request.Temperature),
		topP:             float64(request.TopP),
		maxTokens:        int64(request.MaxTokens),
		frequencyPenalty: float64(request.FrequencyPenalty),
		presencePenalty:  float64(request.PresencePenalty),
		uid:              request.User,
		isStream:         request.Stream,
	}
	if request.Seed != nil {
		req.seed = int64(*request.Seed)
	}
	input, err := json.Marshal(request.Messages)
	if err == nil {
		req.inputMessages = string(input)
	}

	recorder := NewAIMetricsRecorder()
	instrumentedCtx := recorder.Start(ctx, req)

	data := make(map[string]interface{})
	data["ctx"] = instrumentedCtx
	data["request"] = req
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname communityCreateChatCompletionStreamOnExit github.com/sashabaranov/go-openai.communityCreateChatCompletionStreamOnExit
func communityCreateChatCompletionStreamOnExit(call api.CallContext, stream *openai.ChatCompletionStream, err error) {
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

	// For streaming, record the start but actual response data
	// will be collected as the stream is consumed
	response := openaiResponse{}
	recorder.End(ctx, request, response, err)
}
