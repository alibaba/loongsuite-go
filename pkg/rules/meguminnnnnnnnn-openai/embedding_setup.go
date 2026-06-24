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

	"github.com/alibaba/loongsuite-go/pkg/api"
	openai "github.com/meguminnnnnnnnn/go-openai"
	_ "unsafe"
)

//go:linkname communityCreateEmbeddingsOnEnter github.com/meguminnnnnnnnn/go-openai.communityCreateEmbeddingsOnEnter
func communityCreateEmbeddingsOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, conv openai.EmbeddingRequestConverter) {
	if !openaiEnabler.Enable() {
		return
	}
	request := conv.Convert()
	req := openaiRequest{
		operationName: OperationNameEmbeddings,
		modelName:     string(request.Model),
		providerName:  getProviderName(client),
		uid:           request.User,
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

//go:linkname communityCreateEmbeddingsOnExit github.com/meguminnnnnnnnn/go-openai.communityCreateEmbeddingsOnExit
func communityCreateEmbeddingsOnExit(call api.CallContext, resp openai.EmbeddingResponse, err error) {
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
		response.responseModel = string(resp.Model)
		response.usageInputTokens = int64(resp.Usage.PromptTokens)
		response.usageOutputTokens = int64(resp.Usage.CompletionTokens)
		response.usageTotalTokens = int64(resp.Usage.TotalTokens)
		request.inputTokens = response.usageInputTokens
	}

	recorder.End(ctx, request, response, err)
}
