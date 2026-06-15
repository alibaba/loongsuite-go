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

package langchain

import (
	"context"
	"encoding/json"
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

//go:linkname openaiGenerateContentOnEnter github.com/tmc/langchaingo/llms/openai.openaiGenerateContentOnEnter
func openaiGenerateContentOnEnter(call api.CallContext, llm *openai.LLM,
	ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption,
) {
	request := &langChainLLMRequest{
		moduleName:    "unKnown",
		operationName: "chat",
		spanKind:      ai.GenAISpanKindGeneration,
	}
	client := reflect.ValueOf(*llm).FieldByName("client")
	if client.IsValid() && !client.IsNil() {
		if client.Elem().FieldByName("Model").IsValid() {
			request.moduleName = client.Elem().FieldByName("Model").String()
		}
		if client.Elem().FieldByName("baseURL").IsValid() {
			request.serverAddress = client.Elem().FieldByName("baseURL").String()
		}
	}
	LLMBaseOnEnter(call, ctx, request, messages, options...)
}

//go:linkname openaiGenerateContentOnExit github.com/tmc/langchaingo/llms/openai.openaiGenerateContentOnExit
func openaiGenerateContentOnExit(call api.CallContext, resp *llms.ContentResponse, err error) {
	data := call.GetData().(map[string]interface{})
	request := langChainLLMRequest{}
	response := langChainLLMResponse{}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	if err != nil {
		langChainLLMInstrument.End(ctx, request, response, err)
		return
	}
	request = data["request"].(langChainLLMRequest)

	if len(resp.Choices) > 0 {
		var finishReasons []string
		var outputContents []map[string]string
		for _, choice := range resp.Choices {
			finishReasons = append(finishReasons, choice.StopReason)
			outputContents = append(outputContents, map[string]string{
				"role":    "assistant",
				"content": choice.Content,
			})
		}
		response.responseFinishReasons = finishReasons
		if outputJSON, err := json.Marshal(outputContents); err == nil {
			response.output = string(outputJSON)
		}
		if totalTokensAny, ok1 := resp.Choices[0].GenerationInfo["TotalTokens"]; ok1 {
			if totalTokens, ok2 := totalTokensAny.(int); ok2 {
				response.usageOutputTokens = int64(totalTokens)
			}
		}
		if reasoningTokensAny, ok1 := resp.Choices[0].GenerationInfo["ReasoningTokens"]; ok1 {
			if totalTokens, ok2 := reasoningTokensAny.(int); ok2 {
				request.usageInputTokens = int64(totalTokens)
			}
		}
	}

	langChainLLMInstrument.End(ctx, request, response, nil)
}

//go:linkname ollamaGenerateContentOnEnter github.com/tmc/langchaingo/llms/ollama.ollamaGenerateContentOnEnter
func ollamaGenerateContentOnEnter(call api.CallContext, llm *ollama.LLM,
	ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption,
) {
	request := &langChainLLMRequest{
		moduleName:    "unKnown",
		operationName: "chat",
		spanKind:      ai.GenAISpanKindGeneration,
	}
	opt := reflect.ValueOf(*llm).FieldByName("options")
	if opt.IsValid() {
		if opt.FieldByName("model").IsValid() {
			request.moduleName = opt.FieldByName("model").String()
		}
	}
	LLMBaseOnEnter(call, ctx, request, messages, options...)
}

//go:linkname ollamaGenerateContentOnExit github.com/tmc/langchaingo/llms/ollama.ollamaGenerateContentOnExit
func ollamaGenerateContentOnExit(call api.CallContext, resp *llms.ContentResponse, err error) {
	data := call.GetData().(map[string]interface{})
	request := langChainLLMRequest{}
	response := langChainLLMResponse{}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	if err != nil {
		langChainLLMInstrument.End(ctx, request, response, err)
		return
	}
	request = data["request"].(langChainLLMRequest)

	if len(resp.Choices) > 0 {
		var finishReasons []string
		var outputContents []map[string]string
		for _, choice := range resp.Choices {
			finishReasons = append(finishReasons, choice.StopReason)
			outputContents = append(outputContents, map[string]string{
				"role":    "assistant",
				"content": choice.Content,
			})
		}
		response.responseFinishReasons = finishReasons
		// Serialize output messages
		if outputJSON, err := json.Marshal(outputContents); err == nil {
			response.output = string(outputJSON)
		}
		if totalTokensAny, ok1 := resp.Choices[0].GenerationInfo["TotalTokens"]; ok1 {
			if totalTokens, ok2 := totalTokensAny.(int); ok2 {
				response.usageOutputTokens = int64(totalTokens)
			}
		}
	}
	langChainLLMInstrument.End(ctx, request, response, nil)
}

func LLMBaseOnEnter(call api.CallContext,
	ctx context.Context, req *langChainLLMRequest, messages []llms.MessageContent, options ...llms.CallOption,
) {

	llmsOpts := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmsOpts)
	}
	if llmsOpts.Model != "" {
		req.moduleName = llmsOpts.Model
	}
	req.frequencyPenalty = llmsOpts.FrequencyPenalty
	req.presencePenalty = llmsOpts.PresencePenalty
	req.maxTokens = int64(llmsOpts.MaxTokens)
	req.temperature = llmsOpts.Temperature
	req.stopSequences = llmsOpts.StopWords
	req.topK = float64(llmsOpts.TopK)
	req.topP = llmsOpts.TopP
	req.seed = int64(llmsOpts.Seed)

	// Serialize input messages
	if inputJSON, err := json.Marshal(messages); err == nil {
		req.input = string(inputJSON)
	}

	langCtx := langChainLLMInstrument.Start(ctx, *req)
	data := make(map[string]interface{})
	data["ctx"] = langCtx
	data["request"] = *req
	call.SetData(data)
}
