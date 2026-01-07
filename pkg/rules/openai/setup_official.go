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

// Hooks for github.com/openai/openai-go (official OpenAI SDK)

//go:linkname officialNewChatCompletionOnEnter github.com/openai/openai-go/chat/completions.officialNewChatCompletionOnEnter
func officialNewChatCompletionOnEnter(call api.CallContext, ctx context.Context, params interface{}) {
	if !openaiEnabler.Enable() {
		return
	}

	request := openaiRequest{
		operationName: OperationNameChat,
		modelName:     "unknown",
	}

	// Extract parameters using reflection
	paramsVal := reflect.ValueOf(params)
	if paramsVal.Kind() == reflect.Ptr {
		paramsVal = paramsVal.Elem()
	}

	// Extract model name
	if modelField := paramsVal.FieldByName("Model"); modelField.IsValid() {
		if modelField.Kind() == reflect.String {
			request.modelName = modelField.String()
		} else if modelField.Kind() == reflect.Interface || modelField.Kind() == reflect.Ptr {
			if !modelField.IsNil() {
				if strVal := modelField.Elem(); strVal.Kind() == reflect.String {
					request.modelName = strVal.String()
				}
			}
		}
	}

	// Extract temperature
	if tempField := paramsVal.FieldByName("Temperature"); tempField.IsValid() && !tempField.IsNil() {
		if tempField.Kind() == reflect.Float64 {
			request.temperature = tempField.Float()
		} else if tempField.Kind() == reflect.Ptr && tempField.Elem().Kind() == reflect.Float64 {
			request.temperature = tempField.Elem().Float()
		}
	}

	// Extract max_tokens
	if maxTokensField := paramsVal.FieldByName("MaxTokens"); maxTokensField.IsValid() && !maxTokensField.IsNil() {
		if maxTokensField.Kind() == reflect.Int64 {
			request.maxTokens = maxTokensField.Int()
		} else if maxTokensField.Kind() == reflect.Ptr && maxTokensField.Elem().Kind() == reflect.Int64 {
			request.maxTokens = maxTokensField.Elem().Int()
		}
	}

	// Extract top_p
	if topPField := paramsVal.FieldByName("TopP"); topPField.IsValid() && !topPField.IsNil() {
		if topPField.Kind() == reflect.Float64 {
			request.topP = topPField.Float()
		} else if topPField.Kind() == reflect.Ptr && topPField.Elem().Kind() == reflect.Float64 {
			request.topP = topPField.Elem().Float()
		}
	}

	// Extract frequency_penalty
	if freqField := paramsVal.FieldByName("FrequencyPenalty"); freqField.IsValid() && !freqField.IsNil() {
		if freqField.Kind() == reflect.Float64 {
			request.frequencyPenalty = freqField.Float()
		} else if freqField.Kind() == reflect.Ptr && freqField.Elem().Kind() == reflect.Float64 {
			request.frequencyPenalty = freqField.Elem().Float()
		}
	}

	// Extract presence_penalty
	if presField := paramsVal.FieldByName("PresencePenalty"); presField.IsValid() && !presField.IsNil() {
		if presField.Kind() == reflect.Float64 {
			request.presencePenalty = presField.Float()
		} else if presField.Kind() == reflect.Ptr && presField.Elem().Kind() == reflect.Float64 {
			request.presencePenalty = presField.Elem().Float()
		}
	}

	// Extract seed
	if seedField := paramsVal.FieldByName("Seed"); seedField.IsValid() && !seedField.IsNil() {
		if seedField.Kind() == reflect.Int64 {
			request.seed = seedField.Int()
		} else if seedField.Kind() == reflect.Ptr && seedField.Elem().Kind() == reflect.Int64 {
			request.seed = seedField.Elem().Int()
		}
	}

	// Extract stop sequences
	if stopField := paramsVal.FieldByName("Stop"); stopField.IsValid() && !stopField.IsNil() {
		if stopField.Kind() == reflect.Slice {
			stopSeqs := make([]string, 0, stopField.Len())
			for i := 0; i < stopField.Len(); i++ {
				item := stopField.Index(i)
				if item.Kind() == reflect.String {
					stopSeqs = append(stopSeqs, item.String())
				}
			}
			request.stopSequences = stopSeqs
		}
	}

	// Extract messages count
	if messagesField := paramsVal.FieldByName("Messages"); messagesField.IsValid() {
		if messagesField.Kind() == reflect.Slice {
			request.inputMessages = messagesField.Len()
		}
	}

	recorder := NewAIMetricsRecorder()
	instrumentedCtx := recorder.Start(ctx, request)

	// Store context and request in call data
	data := make(map[string]interface{})
	data["ctx"] = instrumentedCtx
	data["request"] = request
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(0, instrumentedCtx)
}

//go:linkname officialNewChatCompletionOnExit github.com/openai/openai-go/chat/completions.officialNewChatCompletionOnExit
func officialNewChatCompletionOnExit(call api.CallContext, resp interface{}, err error) {
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
			if usageField.Kind() == reflect.Ptr {
				usageField = usageField.Elem()
			}
			if promptTokensField := usageField.FieldByName("PromptTokens"); promptTokensField.IsValid() {
				response.usageInputTokens = promptTokensField.Int()
			}
			if completionTokensField := usageField.FieldByName("CompletionTokens"); completionTokensField.IsValid() {
				response.usageOutputTokens = completionTokensField.Int()
			}
			if totalTokensField := usageField.FieldByName("TotalTokens"); totalTokensField.IsValid() {
				response.usageTotalTokens = totalTokensField.Int()
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

//go:linkname officialNewChatCompletionStreamOnEnter github.com/openai/openai-go/chat/completions.officialNewChatCompletionStreamOnEnter
func officialNewChatCompletionStreamOnEnter(call api.CallContext, ctx context.Context, params interface{}) {
	if !openaiEnabler.Enable() {
		return
	}

	request := openaiRequest{
		operationName: OperationNameStream,
		modelName:     "unknown",
	}

	// Extract parameters using reflection (same as non-streaming)
	paramsVal := reflect.ValueOf(params)
	if paramsVal.Kind() == reflect.Ptr {
		paramsVal = paramsVal.Elem()
	}

	// Extract model name
	if modelField := paramsVal.FieldByName("Model"); modelField.IsValid() {
		if modelField.Kind() == reflect.String {
			request.modelName = modelField.String()
		} else if modelField.Kind() == reflect.Interface || modelField.Kind() == reflect.Ptr {
			if !modelField.IsNil() {
				if strVal := modelField.Elem(); strVal.Kind() == reflect.String {
					request.modelName = strVal.String()
				}
			}
		}
	}

	// Extract other parameters (temperature, max_tokens, etc.)
	if tempField := paramsVal.FieldByName("Temperature"); tempField.IsValid() && !tempField.IsNil() {
		if tempField.Kind() == reflect.Float64 {
			request.temperature = tempField.Float()
		} else if tempField.Kind() == reflect.Ptr && tempField.Elem().Kind() == reflect.Float64 {
			request.temperature = tempField.Elem().Float()
		}
	}

	if maxTokensField := paramsVal.FieldByName("MaxTokens"); maxTokensField.IsValid() && !maxTokensField.IsNil() {
		if maxTokensField.Kind() == reflect.Int64 {
			request.maxTokens = maxTokensField.Int()
		} else if maxTokensField.Kind() == reflect.Ptr && maxTokensField.Elem().Kind() == reflect.Int64 {
			request.maxTokens = maxTokensField.Elem().Int()
		}
	}

	if messagesField := paramsVal.FieldByName("Messages"); messagesField.IsValid() {
		if messagesField.Kind() == reflect.Slice {
			request.inputMessages = messagesField.Len()
		}
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

//go:linkname officialNewChatCompletionStreamOnExit github.com/openai/openai-go/chat/completions.officialNewChatCompletionStreamOnExit
func officialNewChatCompletionStreamOnExit(call api.CallContext, stream interface{}, err error) {
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
	recorder.End(ctx, request, response, err)
}
