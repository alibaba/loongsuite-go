// Copyright (c) 2026 Alibaba Group Holding Ltd.
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
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
)

// Executor.Call - Agent execution entry point
//
//go:linkname executorCallOnEnter github.com/tmc/langchaingo/agents.executorCallOnEnter
func executorCallOnEnter(call api.CallContext,
	e *agents.Executor,
	ctx context.Context,
	inputValues map[string]any,
	options ...chains.ChainCallOption,
) {
	if !langChainEnabler.Enable() {
		return
	}
	request := langChainRequest{
		operationName: MAgentExecutor,
		system:        "langchain",
		spanKind:      ai.GenAISpanKindAgent,
		input:         inputValues,
	}
	langCtx := langChainCommonInstrument.Start(ctx, request)
	data := make(map[string]interface{})
	data["ctx"] = langCtx
	call.SetData(data)
}

//go:linkname executorCallOnExit github.com/tmc/langchaingo/agents.executorCallOnExit
func executorCallOnExit(call api.CallContext, result map[string]any, err error) {
	if !langChainEnabler.Enable() {
		return
	}
	dataRaw := call.GetData()
	if dataRaw == nil {
		return
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	request := langChainRequest{
		operationName: MAgentExecutor,
		system:        "langchain",
		output:        result,
	}
	langChainCommonInstrument.End(ctx, request, nil, err)
}
