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

package meguminnnnnnnnn_openai

import (
	"context"
	"encoding/json"

	"github.com/alibaba/loongsuite-go/pkg/api"
	openai "github.com/meguminnnnnnnnn/go-openai"
	_ "unsafe"
)

//go:linkname communityCreateCompletionOnEnter github.com/meguminnnnnnnnn/go-openai.communityCreateCompletionOnEnter
func communityCreateCompletionOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request openai.CompletionRequest) {
	if !openaiEnabler.Enable() {
		return
	}
	req := openaiRequest{
		operationName:    OperationNameTextCompletion,
		modelName:        request.Model,
		providerName:     getProviderName(client),
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
	prompt, err := json.Marshal(request.Prompt)
	if err == nil {
		req.inputMessages = string(prompt)
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

//go:linkname communityCreateCompletionOnExit github.com/meguminnnnnnnnn/go-openai.communityCreateCompletionOnExit
func communityCreateCompletionOnExit(call api.CallContext, resp openai.CompletionResponse, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
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
			if choice.FinishReason != "" {
				response.finishReasons = append(response.finishReasons, string(choice.FinishReason))
			}
			msgs = append(msgs, choice.Text)
		}
		output, err1 := json.Marshal(msgs)
		if err1 == nil {
			response.outputMessages = string(output)
		}
	}

	recorder.End(ctx, request, response, err)
}

//go:linkname communityCreateCompletionStreamOnEnter github.com/meguminnnnnnnnn/go-openai.communityCreateCompletionStreamOnEnter
func communityCreateCompletionStreamOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request openai.CompletionRequest) {
	if !openaiEnabler.Enable() {
		return
	}
	req := openaiRequest{
		operationName:    OperationNameTextCompletion,
		modelName:        request.Model,
		providerName:     getProviderName(client),
		temperature:      float64(request.Temperature),
		topP:             float64(request.TopP),
		maxTokens:        int64(request.MaxTokens),
		frequencyPenalty: float64(request.FrequencyPenalty),
		presencePenalty:  float64(request.PresencePenalty),
		uid:              request.User,
		isStream:         true,
	}
	if request.Seed != nil {
		req.seed = int64(*request.Seed)
	}
	prompt, err := json.Marshal(request.Prompt)
	if err == nil {
		req.inputMessages = string(prompt)
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

//go:linkname communityCreateCompletionStreamOnExit github.com/meguminnnnnnnnn/go-openai.communityCreateCompletionStreamOnExit
func communityCreateCompletionStreamOnExit(call api.CallContext, stream *openai.CompletionStream, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(openaiRequest)
	recorder, _ := data["recorder"].(*AIMetricsRecorder)

	if recorder == nil || ctx == nil {
		return
	}

	if err != nil || stream == nil {
		response := openaiResponse{}
		recorder.End(ctx, request, response, err)
		return
	}

	stream.InitOtelChan()
	go func() {
		var response openaiResponse
		var outputContent string
		for {
			select {
			case resp, ok := <-stream.GetOtelEndChan():
				if !ok {
					if outputContent != "" {
						msgs := []string{outputContent}
						output, err1 := json.Marshal(msgs)
						if err1 == nil {
							response.outputMessages = string(output)
						}
					}
					recorder.End(ctx, request, response, nil)
					return
				}
				resp1, ok1 := resp.(openai.CompletionResponse)
				if ok1 {
					if response.responseID == "" {
						response.responseID = resp1.ID
					}
					for _, r := range resp1.Choices {
						if r.FinishReason != "" {
							response.finishReasons = append(response.finishReasons, string(r.FinishReason))
						}
						outputContent += r.Text
					}
					response.usageInputTokens += int64(resp1.Usage.PromptTokens)
					response.usageOutputTokens += int64(resp1.Usage.CompletionTokens)
					response.usageTotalTokens += int64(resp1.Usage.TotalTokens)
					request.inputTokens = response.usageInputTokens
				}
			}
		}
	}()
}
