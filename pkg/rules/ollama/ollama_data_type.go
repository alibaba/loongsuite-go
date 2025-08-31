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
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

// ollamaRequest represents an Ollama API request
type ollamaRequest struct {
	operationType string        // "chat" or "generate"
	model         string        // Model name
	messages      []api.Message // For chat requests
	prompt        string        // For generate requests

	// Token counts - populated from response
	promptTokens     int
	completionTokens int

	// Streaming flag - true if Stream is nil or *Stream is true
	isStreaming bool
}

// streamingState tracks the state of a streaming response
type streamingState struct {
	// Timing metrics
	startTime      time.Time  // When the request started
	firstTokenTime *time.Time // When the first content chunk arrived (for TTFT)
	endTime        *time.Time // When streaming completed

	// Chunk tracking
	chunkCount    int       // Total number of chunks received
	lastChunkTime time.Time // Time of last chunk (for periodic updates)

	// Content accumulation
	responseBuilder strings.Builder // Accumulates response content

	// Token metrics
	runningTokenCount int     // Running total of tokens generated
	tokenRate         float64 // Tokens per second (calculated)

	// Final metrics from last chunk
	promptEvalCount int           // Input tokens (from final chunk)
	evalCount       int           // Output tokens (accumulated)
	totalDuration   time.Duration // Total generation time
	
	// Cost tracking
	streamingCost *StreamingCostState // Real-time cost accumulation
}

// Helper methods for streamingState

// newStreamingState creates a new streaming state for tracking
func newStreamingState(modelID string) *streamingState {
	state := &streamingState{
		startTime:     time.Now(),
		lastChunkTime: time.Now(),
	}
	
	// Initialize cost tracking if enabled
	calculator := GetCostCalculator()
	if calculator != nil && calculator.IsEnabled() {
		state.streamingCost = calculator.NewStreamingCostState(modelID)
	}
	
	return state
}

// recordChunk processes a streaming chunk and updates state
func (s *streamingState) recordChunk(content string, evalCount int) {
	s.chunkCount++

	// Record TTFT on first content chunk
	if content != "" && s.firstTokenTime == nil {
		now := time.Now()
		s.firstTokenTime = &now
	}

	// Accumulate content
	s.responseBuilder.WriteString(content)

	// Update token count if provided
	if evalCount > 0 {
		s.evalCount = evalCount // This is cumulative in Ollama
		s.runningTokenCount = evalCount
		
		// Update streaming cost if tracking
		if s.streamingCost != nil {
			s.streamingCost.UpdateStreamingCost(evalCount)
		}
	}

	s.lastChunkTime = time.Now()
}

// finalize marks the streaming as complete and calculates final metrics
func (s *streamingState) finalize(promptEvalCount, evalCount int, totalDuration time.Duration) {
	now := time.Now()
	s.endTime = &now
	s.promptEvalCount = promptEvalCount
	s.evalCount = evalCount
	s.totalDuration = totalDuration

	// Calculate token rate if we have duration and tokens
	if totalDuration > 0 && evalCount > 0 {
		s.tokenRate = float64(evalCount) / totalDuration.Seconds()
	}
	
	// Finalize cost tracking with input tokens
	if s.streamingCost != nil && promptEvalCount > 0 {
		s.streamingCost.SetInputTokens(promptEvalCount)
	}
}

// getTTFTMillis returns Time To First Token in milliseconds
func (s *streamingState) getTTFTMillis() int64 {
	if s.firstTokenTime == nil {
		return 0
	}
	return s.firstTokenTime.Sub(s.startTime).Milliseconds()
}

const (
	eventChunkInterval  = 10  // Record event every 10 chunks
	eventTimeIntervalMs = 500 // Record event every 500ms
)

// shouldRecordEvent checks if we should record a span event based on time or chunk count
func (s *streamingState) shouldRecordEvent() bool {
	// Record event every eventChunkInterval chunks or every eventTimeIntervalMs milliseconds
	timeSinceLastEvent := time.Since(s.lastChunkTime)
	return s.chunkCount%eventChunkInterval == 0 || timeSinceLastEvent > eventTimeIntervalMs*time.Millisecond
}

// ollamaResponse represents an Ollama API response
type ollamaResponse struct {
	// Token counts from the response
	promptTokens     int
	completionTokens int

	// Response content
	content string

	// Error if any
	err error

	// Streaming metrics (populated only for streaming responses)
	streamingMetrics *streamingState
	
	// Cost metrics (populated when cost tracking is enabled)
	costMetrics *CostMetrics
}
