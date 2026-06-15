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
	"context"
	"encoding/json"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

var ADKLLMInstrumenter = BuildADKLLMInstrumenter()
var ADKAgentInstrumenter = BuildADKAgentInstrumenter()

// --- Runner.Run hooks: agent workflow span ---

//go:linkname runnerRunOnEnter google.golang.org/adk/runner.runnerRunOnEnter
func runnerRunOnEnter(call api.CallContext, r interface{}, ctx context.Context, userID string, sessionID string, msg *genai.Content, cfg interface{}) {
	if !adkGoEnabler.Enable() {
		return
	}

	request := adkAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		input: map[string]any{
			"session_id": sessionID,
			"user_id":    userID,
		},
	}

	if msg != nil && msg.Parts != nil {
		for _, p := range msg.Parts {
			if p.Text != "" {
				request.input["user_message"] = p.Text
				break
			}
		}
	}

	instrumentedCtx := ADKAgentInstrumenter.Start(ctx, request)
	data := make(map[string]any)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname runnerRunOnExit google.golang.org/adk/runner.runnerRunOnExit
func runnerRunOnExit(call api.CallContext, result any) {
	data, ok := call.GetData().(map[string]any)
	if !ok || data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(adkAgentRequest)
	if ctx == nil {
		return
	}

	ADKAgentInstrumenter.End(ctx, request, nil, nil)
}

// --- Runner.Run hooks for v1.0.0+ (added opts ...RunOption variadic parameter) ---

//go:linkname runnerRunOnEnterV1 google.golang.org/adk/runner.runnerRunOnEnterV1
func runnerRunOnEnterV1(call api.CallContext, r interface{}, ctx context.Context, userID string, sessionID string, msg *genai.Content, cfg interface{}, opts ...interface{}) {
	if !adkGoEnabler.Enable() {
		return
	}

	request := adkAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		input: map[string]any{
			"session_id": sessionID,
			"user_id":    userID,
		},
	}

	if msg != nil && msg.Parts != nil {
		for _, p := range msg.Parts {
			if p.Text != "" {
				request.input["user_message"] = p.Text
				break
			}
		}
	}

	instrumentedCtx := ADKAgentInstrumenter.Start(ctx, request)
	data := make(map[string]any)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname runnerRunOnExitV1 google.golang.org/adk/runner.runnerRunOnExitV1
func runnerRunOnExitV1(call api.CallContext, result any) {
	data, ok := call.GetData().(map[string]any)
	if !ok || data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(adkAgentRequest)
	if ctx == nil {
		return
	}

	ADKAgentInstrumenter.End(ctx, request, nil, nil)
}

// --- geminiModel.generate hooks: non-streaming LLM span with full response ---

//go:linkname geminiGenerateOnEnter google.golang.org/adk/model/gemini.geminiGenerateOnEnter
func geminiGenerateOnEnter(call api.CallContext, m interface{}, ctx context.Context, req *model.LLMRequest) {
	if !adkGoEnabler.Enable() {
		return
	}

	var modelName string
	if llm, ok := m.(interface{ Name() string }); ok {
		modelName = llm.Name()
	}
	if req.Model != "" {
		modelName = req.Model
	}

	request := adkLLMRequest{
		operationName: OperationChat,
		modelName:     modelName,
		isStream:      false,
	}

	if req.Config != nil {
		if req.Config.Temperature != nil {
			request.temperature = float64(*req.Config.Temperature)
		}
		if req.Config.TopP != nil {
			request.topP = float64(*req.Config.TopP)
		}
		if req.Config.TopK != nil {
			request.topK = float64(*req.Config.TopK)
		}
		if req.Config.MaxOutputTokens > 0 {
			request.maxTokens = int64(req.Config.MaxOutputTokens)
		}
		request.stopSequences = req.Config.StopSequences
	}

	if req.Contents != nil {
		input, err := json.Marshal(req.Contents)
		if err == nil {
			request.inputMessages = string(input)
		}
	}

	recorder := NewLLMRecorder()
	instrumentedCtx := recorder.Start(ctx, request)

	data := make(map[string]any)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname geminiGenerateOnExit google.golang.org/adk/model/gemini.geminiGenerateOnExit
func geminiGenerateOnExit(call api.CallContext, resp *model.LLMResponse, err error) {
	data, ok := call.GetData().(map[string]any)
	if !ok || data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(adkLLMRequest)
	recorder, _ := data["recorder"].(*LLMRecorder)
	if recorder == nil || ctx == nil {
		return
	}

	response := adkLLMResponse{}
	if err == nil && resp != nil {
		response.responseModel = resp.ModelVersion
		if resp.FinishReason != "" {
			response.finishReasons = []string{string(resp.FinishReason)}
		}
		if resp.UsageMetadata != nil {
			response.usageInputTokens = int64(resp.UsageMetadata.PromptTokenCount)
			response.usageOutputTokens = int64(resp.UsageMetadata.CandidatesTokenCount)
			request.inputTokens = response.usageInputTokens
		}
		if resp.Content != nil {
			output, jsonErr := json.Marshal(resp.Content)
			if jsonErr == nil {
				response.outputMessages = string(output)
			}
		}
	}

	recorder.End(ctx, request, response, err)
}

// --- geminiModel.generateStream hooks: streaming LLM span ---

//go:linkname geminiGenerateStreamOnEnter google.golang.org/adk/model/gemini.geminiGenerateStreamOnEnter
func geminiGenerateStreamOnEnter(call api.CallContext, m interface{}, ctx context.Context, req *model.LLMRequest) {
	if !adkGoEnabler.Enable() {
		return
	}

	var modelName string
	if llm, ok := m.(interface{ Name() string }); ok {
		modelName = llm.Name()
	}
	if req.Model != "" {
		modelName = req.Model
	}

	request := adkLLMRequest{
		operationName: OperationChat,
		modelName:     modelName,
		isStream:      true,
	}

	if req.Config != nil {
		if req.Config.Temperature != nil {
			request.temperature = float64(*req.Config.Temperature)
		}
		if req.Config.TopP != nil {
			request.topP = float64(*req.Config.TopP)
		}
		if req.Config.TopK != nil {
			request.topK = float64(*req.Config.TopK)
		}
		if req.Config.MaxOutputTokens > 0 {
			request.maxTokens = int64(req.Config.MaxOutputTokens)
		}
		request.stopSequences = req.Config.StopSequences
	}

	if req.Contents != nil {
		input, jsonErr := json.Marshal(req.Contents)
		if jsonErr == nil {
			request.inputMessages = string(input)
		}
	}

	recorder := NewLLMRecorder()
	instrumentedCtx := recorder.Start(ctx, request)

	data := make(map[string]any)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	data["recorder"] = recorder
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname geminiGenerateStreamOnExit google.golang.org/adk/model/gemini.geminiGenerateStreamOnExit
func geminiGenerateStreamOnExit(call api.CallContext, result any) {
	data, ok := call.GetData().(map[string]any)
	if !ok || data == nil {
		return
	}

	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(adkLLMRequest)
	recorder, _ := data["recorder"].(*LLMRecorder)
	if recorder == nil || ctx == nil {
		return
	}

	response := adkLLMResponse{}
	recorder.End(ctx, request, response, nil)
}
