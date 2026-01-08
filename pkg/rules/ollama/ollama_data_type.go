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

type ollamaRequest struct {
	operationType string
	model         string
	messages      []api.Message
	prompt        string

	promptTokens     int
	completionTokens int

	isStreaming bool

	embeddingCount   int
	embeddingDim     int
	modelOperation   string
	downloadProgress float64
	serverAddress    string

	temperature      float64
	maxTokens        int64
	topK             float64
	topP             float64
	frequencyPenalty float64
	presencePenalty  float64
	stopSequences    []string
	seed             int64
	input            string
}

type streamingState struct {
	startTime      time.Time
	firstTokenTime *time.Time
	endTime        *time.Time

	chunkCount    int
	lastChunkTime time.Time

	responseBuilder strings.Builder

	runningTokenCount int
	tokenRate         float64

	promptEvalCount int
	evalCount       int
	totalDuration   time.Duration

	streamingCost *StreamingCostState
}

func newStreamingState(modelID string) *streamingState {
	state := &streamingState{
		startTime:     time.Now(),
		lastChunkTime: time.Now(),
	}

	calculator := costCalculator
	if calculator != nil && calculator.IsEnabled() {
		state.streamingCost = calculator.NewStreamingCostState(modelID)
	}

	return state
}

func (s *streamingState) recordChunk(content string, evalCount int) {
	s.chunkCount++

	if content != "" && s.firstTokenTime == nil {
		now := time.Now()
		s.firstTokenTime = &now
	}

	s.responseBuilder.WriteString(content)

	if evalCount > 0 {
		s.evalCount = evalCount
		s.runningTokenCount = evalCount

		if s.streamingCost != nil {
			s.streamingCost.UpdateStreamingCost(evalCount)
		}
	}

	s.lastChunkTime = time.Now()
}

func (s *streamingState) finalize(promptEvalCount, evalCount int, totalDuration time.Duration) {
	now := time.Now()
	s.endTime = &now
	s.promptEvalCount = promptEvalCount
	s.evalCount = evalCount
	s.totalDuration = totalDuration

	if totalDuration > 0 && evalCount > 0 {
		s.tokenRate = float64(evalCount) / totalDuration.Seconds()
	}

	if s.streamingCost != nil && promptEvalCount > 0 {
		s.streamingCost.SetInputTokens(promptEvalCount)
	}
}

func (s *streamingState) getTTFTMillis() int64 {
	if s.firstTokenTime == nil {
		return 0
	}
	return s.firstTokenTime.Sub(s.startTime).Milliseconds()
}

const (
	eventChunkInterval  = 10
	eventTimeIntervalMs = 500
)

func (s *streamingState) shouldRecordEvent() bool {
	timeSinceLastEvent := time.Since(s.lastChunkTime)
	return s.chunkCount%eventChunkInterval == 0 || timeSinceLastEvent > eventTimeIntervalMs*time.Millisecond
}

type ollamaResponse struct {
	promptTokens     int
	completionTokens int

	content string

	err error

	streamingMetrics *streamingState

	costMetrics *CostMetrics

	embeddings   [][]float64
	modelInfo    map[string]interface{}
	modelList    []interface{}
	pullStatus   string
	pullProgress float64
}
