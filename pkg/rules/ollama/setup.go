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
	_ "unsafe" // Required for go:linkname

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	ollamaapi "github.com/ollama/ollama/api"
)

const (
	contentPreviewMaxLen = 100 // Maximum length for content preview in span events
)

// Helper function to truncate strings for span attributes
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GENERATE API HOOKS

//go:linkname clientGenerateOnEnter github.com/ollama/ollama/api.clientGenerateOnEnter
func clientGenerateOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.GenerateRequest, fn ollamaapi.GenerateResponseFunc) {
	// Detect streaming mode (Stream is nil or *Stream is true)
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)

	// Create request tracking object
	ollamaReq := ollamaRequest{
		operationType: "generate",
		model:         req.Model,
		prompt:        req.Prompt,
		isStreaming:   isStreaming,
	}

	// Start OpenTelemetry span
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)

	// Update context parameter
	call.SetParam(1, ctx)

	// Initialize streaming state if streaming
	var streamState *streamingState
	if isStreaming {
		streamState = newStreamingState(req.Model)
	}

	// CRITICAL: Wrap the callback to capture response data
	var finalResponse ollamaapi.GenerateResponse
	var wrappedFn ollamaapi.GenerateResponseFunc = func(resp ollamaapi.GenerateResponse) error {
		// Handle streaming chunks
		if isStreaming && streamState != nil {
			// Record chunk data
			streamState.recordChunk(resp.Response, resp.EvalCount)

			// Record span events for streaming milestones
			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				// First token event
				if streamState.chunkCount == 1 && resp.Response != "" {
					span.AddEvent("First token received",
						trace.WithAttributes(
							attribute.Int64("gen_ai.response.ttft_ms", streamState.getTTFTMillis()),
							attribute.Int("chunk_number", streamState.chunkCount),
						))
				}

				// Periodic updates (every 10 chunks or 500ms)
				if streamState.shouldRecordEvent() {
					span.AddEvent("Streaming progress",
						trace.WithAttributes(
							attribute.Int("chunk_count", streamState.chunkCount),
							attribute.Int("tokens_generated", streamState.runningTokenCount),
							attribute.String("content_preview", truncateString(streamState.responseBuilder.String(), contentPreviewMaxLen)),
						))
				}

				// Final chunk event
				if resp.Done {
					span.AddEvent("Streaming completed",
						trace.WithAttributes(
							attribute.Int("total_chunks", streamState.chunkCount),
							attribute.Int("total_tokens", resp.EvalCount),
							attribute.Float64("tokens_per_second", streamState.tokenRate),
						))
				}
			}

			// Check if this is the final chunk
			if resp.Done {
				streamState.finalize(resp.PromptEvalCount, resp.EvalCount, resp.TotalDuration)
			}
		}

		// Always update with latest response
		// The final response will have Done=true
		if resp.Done {
			finalResponse = resp
		}

		// Call the original callback if provided
		if fn != nil {
			return fn(resp)
		}
		return nil
	}

	// Replace the callback parameter with our wrapped version
	call.SetParam(3, wrappedFn)

	// Store context and response pointer for OnExit
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
	// Retrieve data stored in OnEnter
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	// Get context from data
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}

	// Get request from data
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}

	// Get streaming state if present
	streamState, isStreaming := data["streamingState"].(*streamingState)

	// Create response object
	ollamaResp := ollamaResponse{
		err: err,
	}

	// Add streaming metrics if this was a streaming request
	if isStreaming && streamState != nil {
		ollamaResp.streamingMetrics = streamState
	}

	// Extract response data if no error
	if err == nil {
		// Get the final response captured by our wrapped callback
		if respPtr, ok := data["finalResponsePtr"].(*ollamaapi.GenerateResponse); ok && respPtr != nil {
			// For streaming, use accumulated content
			if isStreaming && streamState != nil {
				ollamaResp.content = streamState.responseBuilder.String()
				ollamaResp.promptTokens = streamState.promptEvalCount
				ollamaResp.completionTokens = streamState.evalCount
			} else {
				ollamaResp.promptTokens = respPtr.PromptEvalCount
				ollamaResp.completionTokens = respPtr.EvalCount
				ollamaResp.content = respPtr.Response
			}

			// Update request with token counts for the instrumenter
			reqPtr.promptTokens = ollamaResp.promptTokens
			reqPtr.completionTokens = ollamaResp.completionTokens
			
			// Calculate costs if enabled
			calculator := GetCostCalculator()
			if calculator != nil && calculator.IsEnabled() {
				// For streaming, get cost from streaming state
				if isStreaming && streamState != nil && streamState.streamingCost != nil {
					ollamaResp.costMetrics = streamState.streamingCost.GetMetrics()
				} else {
					// For non-streaming, calculate cost directly
					costMetrics, _ := calculator.CalculateCost(
						reqPtr.model,
						ollamaResp.promptTokens,
						ollamaResp.completionTokens,
					)
					ollamaResp.costMetrics = costMetrics
				}
				
				// Record cost in budget tracker
				if ollamaResp.costMetrics != nil && ollamaResp.costMetrics.TotalCost > 0 {
					budgetTracker := GetBudgetTracker()
					if budgetTracker != nil {
						budgetTracker.RecordCost(ollamaResp.costMetrics.TotalCost)
					}
				}
			}
		}
	}

	// End OpenTelemetry span
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}

// CHAT API HOOKS

//go:linkname clientChatOnEnter github.com/ollama/ollama/api.clientChatOnEnter
func clientChatOnEnter(call api.CallContext, c *ollamaapi.Client, ctx context.Context, req *ollamaapi.ChatRequest, fn ollamaapi.ChatResponseFunc) {
	// Detect streaming mode (Stream is nil or *Stream is true)
	isStreaming := req.Stream == nil || (req.Stream != nil && *req.Stream)

	// Create request tracking object
	ollamaReq := ollamaRequest{
		operationType: "chat",
		model:         req.Model,
		messages:      req.Messages,
		isStreaming:   isStreaming,
	}

	// Start OpenTelemetry span
	ctx = ollamaInstrumenter.Start(ctx, ollamaReq)

	// Update context parameter
	call.SetParam(1, ctx)

	// Initialize streaming state if streaming
	var streamState *streamingState
	if isStreaming {
		streamState = newStreamingState(req.Model)
	}

	// CRITICAL: Wrap the callback to capture response data
	var finalResponse ollamaapi.ChatResponse
	var wrappedFn ollamaapi.ChatResponseFunc = func(resp ollamaapi.ChatResponse) error {
		// Handle streaming chunks
		if isStreaming && streamState != nil {
			// Record chunk data (for chat, content is in Message.Content)
			streamState.recordChunk(resp.Message.Content, resp.EvalCount)

			// Record span events for streaming milestones
			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				// First token event
				if streamState.chunkCount == 1 && resp.Message.Content != "" {
					span.AddEvent("First token received",
						trace.WithAttributes(
							attribute.Int64("gen_ai.response.ttft_ms", streamState.getTTFTMillis()),
							attribute.Int("chunk_number", streamState.chunkCount),
						))
				}

				// Periodic updates (every 10 chunks or 500ms)
				if streamState.shouldRecordEvent() {
					span.AddEvent("Streaming progress",
						trace.WithAttributes(
							attribute.Int("chunk_count", streamState.chunkCount),
							attribute.Int("tokens_generated", streamState.runningTokenCount),
							attribute.String("content_preview", truncateString(streamState.responseBuilder.String(), contentPreviewMaxLen)),
						))
				}

				// Final chunk event
				if resp.Done {
					span.AddEvent("Streaming completed",
						trace.WithAttributes(
							attribute.Int("total_chunks", streamState.chunkCount),
							attribute.Int("total_tokens", resp.EvalCount),
							attribute.Float64("tokens_per_second", streamState.tokenRate),
						))
				}
			}

			// Check if this is the final chunk
			if resp.Done {
				streamState.finalize(resp.PromptEvalCount, resp.EvalCount, resp.TotalDuration)
			}
		}

		// Always update with latest response
		// The final response will have Done=true
		if resp.Done {
			finalResponse = resp
		}

		// Call the original callback if provided
		if fn != nil {
			return fn(resp)
		}
		return nil
	}

	// Replace the callback parameter with our wrapped version
	call.SetParam(3, wrappedFn)

	// Store context and response pointer for OnExit
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
	// Retrieve data stored in OnEnter
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	// Get context from data
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}

	// Get request from data
	reqPtr, ok := data["request"].(*ollamaRequest)
	if !ok || reqPtr == nil {
		return
	}

	// Get streaming state if present
	streamState, isStreaming := data["streamingState"].(*streamingState)

	// Create response object
	ollamaResp := ollamaResponse{
		err: err,
	}

	// Add streaming metrics if this was a streaming request
	if isStreaming && streamState != nil {
		ollamaResp.streamingMetrics = streamState
	}

	// Extract response data if no error
	if err == nil {
		// Get the final response captured by our wrapped callback
		if respPtr, ok := data["finalResponsePtr"].(*ollamaapi.ChatResponse); ok && respPtr != nil {
			// For streaming, use accumulated content
			if isStreaming && streamState != nil {
				ollamaResp.content = streamState.responseBuilder.String()
				ollamaResp.promptTokens = streamState.promptEvalCount
				ollamaResp.completionTokens = streamState.evalCount
			} else {
				// Token counts are in embedded Metrics struct
				ollamaResp.promptTokens = respPtr.PromptEvalCount
				ollamaResp.completionTokens = respPtr.EvalCount
				ollamaResp.content = respPtr.Message.Content
			}

			// Update request with token counts for the instrumenter
			reqPtr.promptTokens = ollamaResp.promptTokens
			reqPtr.completionTokens = ollamaResp.completionTokens
			
			// Calculate costs if enabled
			calculator := GetCostCalculator()
			if calculator != nil && calculator.IsEnabled() {
				// For streaming, get cost from streaming state
				if isStreaming && streamState != nil && streamState.streamingCost != nil {
					ollamaResp.costMetrics = streamState.streamingCost.GetMetrics()
				} else {
					// For non-streaming, calculate cost directly
					costMetrics, _ := calculator.CalculateCost(
						reqPtr.model,
						ollamaResp.promptTokens,
						ollamaResp.completionTokens,
					)
					ollamaResp.costMetrics = costMetrics
				}
				
				// Record cost in budget tracker
				if ollamaResp.costMetrics != nil && ollamaResp.costMetrics.TotalCost > 0 {
					budgetTracker := GetBudgetTracker()
					if budgetTracker != nil {
						budgetTracker.RecordCost(ollamaResp.costMetrics.TotalCost)
					}
				}
			}
		}
	}

	// End OpenTelemetry span
	ollamaInstrumenter.End(ctx, *reqPtr, ollamaResp, err)
}
