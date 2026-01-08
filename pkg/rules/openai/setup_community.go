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
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

// Hooks for github.com/sashabaranov/go-openai (community SDK)
// Note: This also covers forks like github.com/meguminnnnnnnnn/go-openai

//go:linkname communityCreateChatCompletionOnEnter github.com/sashabaranov/go-openai.communityCreateChatCompletionOnEnter
func communityCreateChatCompletionOnEnter(call api.CallContext, client interface{}, ctx context.Context, request interface{}) {
	if !openaiEnabler.Enable() {
		return
	}

	req := openaiRequest{
		operationName: OperationNameChat,
		modelName:     "unknown",
	}

	// Extract request parameters using reflection
	reqVal := reflect.ValueOf(request)
	if reqVal.Kind() == reflect.Ptr {
		reqVal = reqVal.Elem()
	}

	// Extract model
	if modelField := reqVal.FieldByName("Model"); modelField.IsValid() && modelField.Kind() == reflect.String {
		req.modelName = modelField.String()
	}

	// Extract temperature (can be float32 or float64)
	if tempField := reqVal.FieldByName("Temperature"); tempField.IsValid() {
		if tempField.Kind() == reflect.Float64 {
			req.temperature = tempField.Float()
		} else if tempField.Kind() == reflect.Float32 {
			req.temperature = float64(tempField.Float())
		}
	}

	// Extract max_tokens
	if maxTokensField := reqVal.FieldByName("MaxTokens"); maxTokensField.IsValid() && maxTokensField.Kind() == reflect.Int {
		req.maxTokens = int64(maxTokensField.Int())
	}

	// Extract top_p (can be float32 or float64)
	if topPField := reqVal.FieldByName("TopP"); topPField.IsValid() {
		if topPField.Kind() == reflect.Float64 {
			req.topP = topPField.Float()
		} else if topPField.Kind() == reflect.Float32 {
			req.topP = float64(topPField.Float())
		}
	}

	// Extract frequency_penalty (can be float32 or float64)
	if freqField := reqVal.FieldByName("FrequencyPenalty"); freqField.IsValid() {
		if freqField.Kind() == reflect.Float64 {
			req.frequencyPenalty = freqField.Float()
		} else if freqField.Kind() == reflect.Float32 {
			req.frequencyPenalty = float64(freqField.Float())
		}
	}

	// Extract presence_penalty (can be float32 or float64)
	if presField := reqVal.FieldByName("PresencePenalty"); presField.IsValid() {
		if presField.Kind() == reflect.Float64 {
			req.presencePenalty = presField.Float()
		} else if presField.Kind() == reflect.Float32 {
			req.presencePenalty = float64(presField.Float())
		}
	}

	// Extract seed
	if seedField := reqVal.FieldByName("Seed"); seedField.IsValid() && !seedField.IsNil() {
		if seedField.Kind() == reflect.Ptr && seedField.Elem().Kind() == reflect.Int {
			req.seed = seedField.Elem().Int()
		}
	}

	// Extract stop sequences
	if stopField := reqVal.FieldByName("Stop"); stopField.IsValid() {
		if stopField.Kind() == reflect.Slice {
			stopSeqs := make([]string, 0, stopField.Len())
			for i := 0; i < stopField.Len(); i++ {
				item := stopField.Index(i)
				if item.Kind() == reflect.String {
					stopSeqs = append(stopSeqs, item.String())
				}
			}
			req.stopSequences = stopSeqs
		}
	}

	// Extract messages count
	if messagesField := reqVal.FieldByName("Messages"); messagesField.IsValid() {
		if messagesField.Kind() == reflect.Slice {
			req.inputMessages = messagesField.Len()
		}
	}

	// Extract server address from client
	clientVal := reflect.ValueOf(client)
	if clientVal.Kind() == reflect.Ptr {
		clientVal = clientVal.Elem()
	}
	if configField := clientVal.FieldByName("config"); configField.IsValid() {
		if baseURLField := configField.FieldByName("BaseURL"); baseURLField.IsValid() && baseURLField.Kind() == reflect.String {
			req.serverAddress = baseURLField.String()
		}
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
func communityCreateChatCompletionOnExit(call api.CallContext, resp interface{}, err error) {
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
		respVal := reflect.ValueOf(resp)
		if respVal.Kind() == reflect.Ptr {
			respVal = respVal.Elem()
		}

		// Extract response ID
		if idField := respVal.FieldByName("ID"); idField.IsValid() && idField.Kind() == reflect.String {
			response.responseID = idField.String()
		}

		// Extract model
		if modelField := respVal.FieldByName("Model"); modelField.IsValid() && modelField.Kind() == reflect.String {
			response.responseModel = modelField.String()
		}

		// Extract usage
		if usageField := respVal.FieldByName("Usage"); usageField.IsValid() {
			if promptTokensField := usageField.FieldByName("PromptTokens"); promptTokensField.IsValid() {
				response.usageInputTokens = int64(promptTokensField.Int())
				// Also update the request with input tokens for the instrumenter
				request.inputTokens = response.usageInputTokens
			}
			if completionTokensField := usageField.FieldByName("CompletionTokens"); completionTokensField.IsValid() {
				response.usageOutputTokens = int64(completionTokensField.Int())
			}
			if totalTokensField := usageField.FieldByName("TotalTokens"); totalTokensField.IsValid() {
				response.usageTotalTokens = int64(totalTokensField.Int())
			}
		}

		// Extract choices and finish reasons
		if choicesField := respVal.FieldByName("Choices"); choicesField.IsValid() && choicesField.Kind() == reflect.Slice {
			finishReasons := make([]string, 0)
			for i := 0; i < choicesField.Len(); i++ {
				choice := choicesField.Index(i)
				if finishReasonField := choice.FieldByName("FinishReason"); finishReasonField.IsValid() {
					if finishReasonField.Kind() == reflect.String {
						finishReasons = append(finishReasons, finishReasonField.String())
					}
				}
				// Extract content from first choice
				if i == 0 {
					if messageField := choice.FieldByName("Message"); messageField.IsValid() {
						if contentField := messageField.FieldByName("Content"); contentField.IsValid() && contentField.Kind() == reflect.String {
							response.outputContent = contentField.String()
						}
					}
				}
			}
			response.responseFinishReasons = finishReasons
		}
	}

	recorder.End(ctx, request, response, err)
}

//go:linkname communityCreateChatCompletionStreamOnEnter github.com/sashabaranov/go-openai.communityCreateChatCompletionStreamOnEnter
func communityCreateChatCompletionStreamOnEnter(call api.CallContext, client interface{}, ctx context.Context, request interface{}) {
	if !openaiEnabler.Enable() {
		return
	}

	req := openaiRequest{
		operationName: OperationNameStream,
		modelName:     "unknown",
	}

	// Extract request parameters (same as non-streaming)
	reqVal := reflect.ValueOf(request)
	if reqVal.Kind() == reflect.Ptr {
		reqVal = reqVal.Elem()
	}

	if modelField := reqVal.FieldByName("Model"); modelField.IsValid() && modelField.Kind() == reflect.String {
		req.modelName = modelField.String()
	}

	// Extract temperature (can be float32 or float64)
	if tempField := reqVal.FieldByName("Temperature"); tempField.IsValid() {
		if tempField.Kind() == reflect.Float64 {
			req.temperature = tempField.Float()
		} else if tempField.Kind() == reflect.Float32 {
			req.temperature = float64(tempField.Float())
		}
	}

	if maxTokensField := reqVal.FieldByName("MaxTokens"); maxTokensField.IsValid() && maxTokensField.Kind() == reflect.Int {
		req.maxTokens = int64(maxTokensField.Int())
	}

	if messagesField := reqVal.FieldByName("Messages"); messagesField.IsValid() {
		if messagesField.Kind() == reflect.Slice {
			req.inputMessages = messagesField.Len()
		}
	}

	// Extract server address from client
	clientVal := reflect.ValueOf(client)
	if clientVal.Kind() == reflect.Ptr {
		clientVal = clientVal.Elem()
	}
	if configField := clientVal.FieldByName("config"); configField.IsValid() {
		if baseURLField := configField.FieldByName("BaseURL"); baseURLField.IsValid() && baseURLField.Kind() == reflect.String {
			req.serverAddress = baseURLField.String()
		}
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
func communityCreateChatCompletionStreamOnExit(call api.CallContext, stream interface{}, err error) {
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
