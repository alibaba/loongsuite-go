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
	_ "unsafe"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	ollamaapi "github.com/ollama/ollama/api"
)


//go:linkname clientGenerateOnEnter github.com/ollama/ollama/api.clientGenerateOnEnter
func clientGenerateOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.GenerateRequest, fn ollamaapi.GenerateResponseFunc) {
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)
	ollamaReq := ollamaRequest{
		operationType: "generate",
		model:         req.Model,
		prompt:        req.Prompt,
		isStreaming:   isStreaming,
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

			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				if streamState.chunkCount == 1 && resp.Response != "" {
					span.AddEvent("First token received",
						trace.WithAttributes(
							attribute.Int64("gen_ai.response.ttft_ms", streamState.getTTFTMillis()),
							attribute.Int("chunk_number", streamState.chunkCount),
						))
				}

				if streamState.shouldRecordEvent() {
					span.AddEvent("Streaming progress",
						trace.WithAttributes(
							attribute.Int("chunk_count", streamState.chunkCount),
							attribute.Int("tokens_generated", streamState.runningTokenCount),
							attribute.String("content_preview", streamState.responseBuilder.String()),
						))
				}

				if resp.Done {
					span.AddEvent("Streaming completed",
						trace.WithAttributes(
							attribute.Int("total_chunks", streamState.chunkCount),
							attribute.Int("total_tokens", resp.EvalCount),
							attribute.Float64("tokens_per_second", streamState.tokenRate),
						))
				}
			}

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
			
			calculator := costCalculator
			if calculator != nil && calculator.IsEnabled() {
				if isStreaming && streamState != nil && streamState.streamingCost != nil {
					ollamaResp.costMetrics = streamState.streamingCost.GetMetrics()
				} else {
					costMetrics, _ := calculator.CalculateCost(
						reqPtr.model,
						ollamaResp.promptTokens,
						ollamaResp.completionTokens,
					)
					ollamaResp.costMetrics = costMetrics
				}
				
				if ollamaResp.costMetrics != nil && ollamaResp.costMetrics.TotalCost > 0 {
					budgetTracker := budgetTracker
					if budgetTracker != nil {
						budgetTracker.RecordCost(ollamaResp.costMetrics.TotalCost)
					}
				}
			}
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}


//go:linkname clientChatOnEnter github.com/ollama/ollama/api.clientChatOnEnter
func clientChatOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.ChatRequest, fn ollamaapi.ChatResponseFunc) {
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)
	ollamaReq := ollamaRequest{
		operationType: "chat",
		model:         req.Model,
		messages:      req.Messages,
		isStreaming:   isStreaming,
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

			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				if streamState.chunkCount == 1 && resp.Message.Content != "" {
					span.AddEvent("First token received",
						trace.WithAttributes(
							attribute.Int64("gen_ai.response.ttft_ms", streamState.getTTFTMillis()),
							attribute.Int("chunk_number", streamState.chunkCount),
						))
				}

				if streamState.shouldRecordEvent() {
					span.AddEvent("Streaming progress",
						trace.WithAttributes(
							attribute.Int("chunk_count", streamState.chunkCount),
							attribute.Int("tokens_generated", streamState.runningTokenCount),
							attribute.String("content_preview", streamState.responseBuilder.String()),
						))
				}

				if resp.Done {
					span.AddEvent("Streaming completed",
						trace.WithAttributes(
							attribute.Int("total_chunks", streamState.chunkCount),
							attribute.Int("total_tokens", resp.EvalCount),
							attribute.Float64("tokens_per_second", streamState.tokenRate),
						))
				}
			}

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
			
			calculator := costCalculator
			if calculator != nil && calculator.IsEnabled() {
				if isStreaming && streamState != nil && streamState.streamingCost != nil {
					ollamaResp.costMetrics = streamState.streamingCost.GetMetrics()
				} else {
					costMetrics, _ := calculator.CalculateCost(
						reqPtr.model,
						ollamaResp.promptTokens,
						ollamaResp.completionTokens,
					)
					ollamaResp.costMetrics = costMetrics
				}
				
				if ollamaResp.costMetrics != nil && ollamaResp.costMetrics.TotalCost > 0 {
					budgetTracker := budgetTracker
					if budgetTracker != nil {
						budgetTracker.RecordCost(ollamaResp.costMetrics.TotalCost)
					}
				}
			}
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientEmbedOnEnter github.com/ollama/ollama/api.clientEmbedOnEnter
func clientEmbedOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.EmbedRequest) {
	var promptStr string
	if s, ok := req.Input.(string); ok {
		promptStr = s
	}
	ollamaReq := ollamaRequest{
		operationType:  "embed",
		model:          req.Model,
		prompt:         promptStr,
		embeddingCount: 1,
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
			costMetrics := calculateEmbeddingCost(reqPtr.model, 1, reqPtr.embeddingDim)
			ollamaResp.costMetrics = costMetrics
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientEmbeddingsOnEnter github.com/ollama/ollama/api.clientEmbeddingsOnEnter
func clientEmbeddingsOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.EmbeddingRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "embeddings",
		model:          req.Model,
		prompt:         req.Prompt,
		embeddingCount: 1,
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
			costMetrics := calculateEmbeddingCost(reqPtr.model, 1, reqPtr.embeddingDim)
			ollamaResp.costMetrics = costMetrics
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
	}
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)
	call.SetParam(1, ctx)
	var wrappedFn ollamaapi.PullProgressFunc = func(progress ollamaapi.ProgressResponse) error {
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			if progress.Total > 0 {
				progressPct := float64(progress.Completed) / float64(progress.Total) * 100
				span.AddEvent("Pull progress",
					trace.WithAttributes(
						attribute.String("status", progress.Status),
						attribute.Float64("progress_percentage", progressPct),
						attribute.Int64("bytes_completed", progress.Completed),
						attribute.Int64("bytes_total", progress.Total),
					))
			}
		}
		if fn != nil {
			return fn(progress)
		}
		return nil
	}
	call.SetParam(3, wrappedFn)
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
		err:        err,
		pullStatus: "completed",
	}
	if err != nil {
		ollamaResp.pullStatus = "failed"
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientListOnEnter github.com/ollama/ollama/api.clientListOnEnter
func clientListOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context) {
	ollamaReq := ollamaRequest{
		operationType:  "list",
		modelOperation: "list",
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
	ollamaResp := ollamaResponse{}
	ollamaResp.err = err
	if err == nil && resp != nil {
		modelList := make([]interface{}, 0, len(resp.Models))
		for _, m := range resp.Models {
			modelList = append(modelList, map[string]interface{}{
				"name":       m.Name,
				"size":       m.Size,
				"modified_at": m.ModifiedAt,
			})
		}
		ollamaResp.modelList = modelList
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientShowOnEnter github.com/ollama/ollama/api.clientShowOnEnter
func clientShowOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.ShowRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "show",
		model:          req.Model,
		modelOperation: "show",
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
	ollamaResp := ollamaResponse{}
	ollamaResp.err = err
	if err == nil && resp != nil {
		ollamaResp.modelInfo = map[string]interface{}{
			"family":         resp.Details.Family,
			"parameter_size": resp.Details.ParameterSize,
			"format":         resp.Details.Format,
			"families":       resp.Details.Families,
		}
	}
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

//go:linkname clientDeleteOnEnter github.com/ollama/ollama/api.clientDeleteOnEnter
func clientDeleteOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.DeleteRequest) {
	ollamaReq := ollamaRequest{
		operationType:  "delete",
		model:          req.Model,
		modelOperation: "delete",
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
