// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ollama

import (
	"context"
	"net/url"
	"reflect"
	"unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	ollamaapi "github.com/ollama/ollama/api"
)

func extractServerAddress(c *ollamaapi.Client) string {
	if c == nil {
		return ""
	}
	clientValue := reflect.ValueOf(c).Elem()
	baseField := clientValue.FieldByName("base")
	if baseField.IsValid() && !baseField.IsNil() {
		baseURL := reflect.NewAt(baseField.Type(), unsafe.Pointer(baseField.UnsafeAddr())).Elem().Interface().(*url.URL)
		if baseURL != nil {
			return baseURL.Host
		}
	}
	return ""
}

func extractOptionsFromMap(opts map[string]interface{}) (
	temperature float64,
	maxTokens int64,
	topK float64,
	topP float64,
	frequencyPenalty float64,
	presencePenalty float64,
	stopSequences []string,
	seed int64,
) {
	if opts == nil {
		return
	}

	if val, ok := opts["temperature"]; ok {
		switch v := val.(type) {
		case float64:
			temperature = v
		case float32:
			temperature = float64(v)
		}
	}

	if val, ok := opts["num_predict"]; ok {
		switch v := val.(type) {
		case float64:
			maxTokens = int64(v)
		case int:
			maxTokens = int64(v)
		case int64:
			maxTokens = v
		}
	}

	if val, ok := opts["top_k"]; ok {
		switch v := val.(type) {
		case float64:
			topK = v
		case int:
			topK = float64(v)
		case int64:
			topK = float64(v)
		}
	}

	if val, ok := opts["top_p"]; ok {
		switch v := val.(type) {
		case float64:
			topP = v
		case float32:
			topP = float64(v)
		}
	}

	if val, ok := opts["frequency_penalty"]; ok {
		switch v := val.(type) {
		case float64:
			frequencyPenalty = v
		case float32:
			frequencyPenalty = float64(v)
		}
	}

	if val, ok := opts["presence_penalty"]; ok {
		switch v := val.(type) {
		case float64:
			presencePenalty = v
		case float32:
			presencePenalty = float64(v)
		}
	}

	if val, ok := opts["stop"]; ok {
		if arr, ok := val.([]interface{}); ok {
			stopSequences = make([]string, 0, len(arr))
			for _, v := range arr {
				if str, ok := v.(string); ok {
					stopSequences = append(stopSequences, str)
				}
			}
		} else if arr, ok := val.([]string); ok {
			stopSequences = arr
		}
	}

	if val, ok := opts["seed"]; ok {
		switch v := val.(type) {
		case float64:
			seed = int64(v)
		case int:
			seed = int64(v)
		case int64:
			seed = v
		}
	}

	return
}


//go:linkname clientGenerateOnEnter github.com/ollama/ollama/api.clientGenerateOnEnter
func clientGenerateOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.GenerateRequest, fn ollamaapi.GenerateResponseFunc) {
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)
	temp, maxTok, tk, tp, fp, pp, stop, seed := extractOptionsFromMap(req.Options)

	ollamaReq := ollamaRequest{
		operationType:    "generate",
		model:            req.Model,
		prompt:           req.Prompt,
		isStreaming:      isStreaming,
		serverAddress:    extractServerAddress(c),
		temperature:      temp,
		maxTokens:        maxTok,
		topK:             tk,
		topP:             tp,
		frequencyPenalty: fp,
		presencePenalty:  pp,
		stopSequences:    stop,
		seed:             seed,
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	var streamState *streamingState
	if isStreaming {
		streamState = newStreamingState(req.Model)
	}
	var finalResponse ollamaapi.GenerateResponse
	var wrappedFn ollamaapi.GenerateResponseFunc = func(resp ollamaapi.GenerateResponse) error {
		if isStreaming && streamState != nil {
			streamState.recordChunk(resp.Response, resp.EvalCount)

			if resp.Done {
				streamState.finalize(resp.PromptEvalCount, resp.EvalCount, resp.TotalDuration)
			}
		}

		if resp.Done {
			finalResponse = resp
		}

		if fn != nil {
			return fn(resp)
		}
		return nil
	}
	call.SetParam(3, wrappedFn)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	data["finalResponsePtr"] = &finalResponse
	if isStreaming {
		data["streamingState"] = streamState
	}
	call.SetData(data)
}

//go:linkname clientGenerateOnExit github.com/ollama/ollama/api.clientGenerateOnExit
func clientGenerateOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	streamState, isStreaming := data["streamingState"].(*streamingState)
	ollamaResp := ollamaResponse{
		err: err,
	}
	if isStreaming && streamState != nil {
		ollamaResp.streamingMetrics = streamState
	}
	if err == nil {
		if respPtr, ok := data["finalResponsePtr"].(*ollamaapi.GenerateResponse); ok && respPtr != nil {
			if isStreaming && streamState != nil {
				ollamaResp.content = streamState.responseBuilder.String()
				ollamaResp.promptTokens = streamState.promptEvalCount
				ollamaResp.completionTokens = streamState.evalCount
			} else {
				ollamaResp.promptTokens = respPtr.PromptEvalCount
				ollamaResp.completionTokens = respPtr.EvalCount
				ollamaResp.content = respPtr.Response
			}

			reqPtr.promptTokens = ollamaResp.promptTokens
			reqPtr.completionTokens = ollamaResp.completionTokens
		}
	}
	// Set TTFT in context for metrics if streaming
	if isStreaming && streamState != nil && streamState.firstTokenTime != nil {
		ctx = context.WithValue(ctx, ai.TimeToFirstTokenKey{}, *streamState.firstTokenTime)
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}


//go:linkname clientChatOnEnter github.com/ollama/ollama/api.clientChatOnEnter
func clientChatOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.ChatRequest, fn ollamaapi.ChatResponseFunc) {
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)
	temp, maxTok, tk, tp, fp, pp, stop, seed := extractOptionsFromMap(req.Options)

	ollamaReq := ollamaRequest{
		operationType:    "chat",
		model:            req.Model,
		messages:         req.Messages,
		isStreaming:      isStreaming,
		serverAddress:    extractServerAddress(c),
		temperature:      temp,
		maxTokens:        maxTok,
		topK:             tk,
		topP:             tp,
		frequencyPenalty: fp,
		presencePenalty:  pp,
		stopSequences:    stop,
		seed:             seed,
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	var streamState *streamingState
	if isStreaming {
		streamState = newStreamingState(req.Model)
	}
	var finalResponse ollamaapi.ChatResponse
	var wrappedFn ollamaapi.ChatResponseFunc = func(resp ollamaapi.ChatResponse) error {
		if isStreaming && streamState != nil {
			streamState.recordChunk(resp.Message.Content, resp.EvalCount)

			if resp.Done {
				streamState.finalize(resp.PromptEvalCount, resp.EvalCount, resp.TotalDuration)
			}
		}

		if resp.Done {
			finalResponse = resp
		}

		if fn != nil {
			return fn(resp)
		}
		return nil
	}
	call.SetParam(3, wrappedFn)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	data["finalResponsePtr"] = &finalResponse
	if isStreaming {
		data["streamingState"] = streamState
	}
	call.SetData(data)
}

//go:linkname clientChatOnExit github.com/ollama/ollama/api.clientChatOnExit
func clientChatOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	streamState, isStreaming := data["streamingState"].(*streamingState)
	ollamaResp := ollamaResponse{
		err: err,
	}
	if isStreaming && streamState != nil {
		ollamaResp.streamingMetrics = streamState
	}
	if err == nil {
		if respPtr, ok := data["finalResponsePtr"].(*ollamaapi.ChatResponse); ok && respPtr != nil {
			if isStreaming && streamState != nil {
				ollamaResp.content = streamState.responseBuilder.String()
				ollamaResp.promptTokens = streamState.promptEvalCount
				ollamaResp.completionTokens = streamState.evalCount
			} else {
					ollamaResp.promptTokens = respPtr.PromptEvalCount
				ollamaResp.completionTokens = respPtr.EvalCount
				ollamaResp.content = respPtr.Message.Content
			}

			reqPtr.promptTokens = ollamaResp.promptTokens
			reqPtr.completionTokens = ollamaResp.completionTokens
		}
	}
	// Set TTFT in context for metrics if streaming
	if isStreaming && streamState != nil && streamState.firstTokenTime != nil {
		ctx = context.WithValue(ctx, ai.TimeToFirstTokenKey{}, *streamState.firstTokenTime)
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientEmbedOnEnter github.com/ollama/ollama/api.clientEmbedOnEnter
func clientEmbedOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.EmbedRequest) {
	var promptStr string
	if s, ok := req.Input.(string); ok {
		promptStr = s
	}
	temp, maxTok, tk, tp, fp, pp, stop, seed := extractOptionsFromMap(req.Options)

	ollamaReq := ollamaRequest{
		operationType:    "embed",
		model:            req.Model,
		prompt:           promptStr,
		embeddingCount:   1,
		serverAddress:    extractServerAddress(c),
		temperature:      temp,
		maxTokens:        maxTok,
		topK:             tk,
		topP:             tp,
		frequencyPenalty: fp,
		presencePenalty:  pp,
		stopSequences:    stop,
		seed:             seed,
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientEmbedOnExit github.com/ollama/ollama/api.clientEmbedOnExit
func clientEmbedOnExit(call api.CallContext, resp *ollamaapi.EmbedResponse, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{}
	ollamaResp.err = err
	if err == nil && resp != nil {
		if len(resp.Embeddings) > 0 {
			embeddings := make([][]float64, len(resp.Embeddings))
			for i, emb := range resp.Embeddings {
				embeddings[i] = make([]float64, len(emb))
				for j, v := range emb {
					embeddings[i][j] = float64(v)
				}
			}
			ollamaResp.embeddings = embeddings
			reqPtr.embeddingDim = len(resp.Embeddings[0])
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientEmbeddingsOnEnter github.com/ollama/ollama/api.clientEmbeddingsOnEnter
func clientEmbeddingsOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.EmbeddingRequest) {
	temp, maxTok, tk, tp, fp, pp, stop, seed := extractOptionsFromMap(req.Options)

	ollamaReq := ollamaRequest{
		operationType:    "embeddings",
		model:            req.Model,
		prompt:           req.Prompt,
		embeddingCount:   1,
		serverAddress:    extractServerAddress(c),
		temperature:      temp,
		maxTokens:        maxTok,
		topK:             tk,
		topP:             tp,
		frequencyPenalty: fp,
		presencePenalty:  pp,
		stopSequences:    stop,
		seed:             seed,
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientEmbeddingsOnExit github.com/ollama/ollama/api.clientEmbeddingsOnExit
func clientEmbeddingsOnExit(call api.CallContext, resp *ollamaapi.EmbeddingResponse, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{}
	ollamaResp.err = err
	if err == nil && resp != nil {
		if len(resp.Embedding) > 0 {
			reqPtr.embeddingDim = len(resp.Embedding)
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientPullOnEnter github.com/ollama/ollama/api.clientPullOnEnter
func clientPullOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.PullRequest, fn ollamaapi.PullProgressFunc) {
	ollamaReq := ollamaRequest{
		operationType:  "pull",
		model:          req.Model,
		modelOperation: "pull",
		serverAddress:  extractServerAddress(c),
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientPullOnExit github.com/ollama/ollama/api.clientPullOnExit
func clientPullOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{
		err: err,
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientListOnEnter github.com/ollama/ollama/api.clientListOnEnter
func clientListOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context) {
	ollamaReq := ollamaRequest{
		operationType:  "list",
		modelOperation: "list",
		serverAddress:  extractServerAddress(c),
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientListOnExit github.com/ollama/ollama/api.clientListOnExit
func clientListOnExit(call api.CallContext, resp *ollamaapi.ListResponse, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{
		err: err,
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientShowOnEnter github.com/ollama/ollama/api.clientShowOnEnter
func clientShowOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.ShowRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "show",
		model:          req.Model,
		modelOperation: "show",
		serverAddress:  extractServerAddress(c),
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientShowOnExit github.com/ollama/ollama/api.clientShowOnExit
func clientShowOnExit(call api.CallContext, resp *ollamaapi.ShowResponse, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{
		err: err,
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientDeleteOnEnter github.com/ollama/ollama/api.clientDeleteOnEnter
func clientDeleteOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.DeleteRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "delete",
		model:          req.Model,
		modelOperation: "delete",
		serverAddress:  extractServerAddress(c),
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	call.SetData(data)
}

//go:linkname clientDeleteOnExit github.com/ollama/ollama/api.clientDeleteOnExit
func clientDeleteOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{
		err: err,
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientCopyOnEnter github.com/ollama/ollama/api.clientCopyOnEnter
func clientCopyOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.CopyRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "copy",
		model:          req.Source,
		modelOperation: "copy",
		serverAddress:  extractServerAddress(c),
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["request"] = &ollamaReq
	data["destination"] = req.Destination
	call.SetData(data)
}

//go:linkname clientCopyOnExit github.com/ollama/ollama/api.clientCopyOnExit
func clientCopyOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}
	ollamaResp := ollamaResponse{
		err: err,
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}
