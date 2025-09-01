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

package eino

import (
	"os"
	"sync"

	"github.com/cloudwego/eino/schema"
)

type einoInnerEnabler struct {
	enabled bool
}

func (l einoInnerEnabler) Enable() bool {
	return l.enabled
}

var einoEnabler = einoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_EINO_ENABLED") != "false"}

type (
	promptRequestKey    struct{}
	llmRequestKey       struct{}
	embeddingRequestKey struct{}
	indexerRequestKey   struct{}
	retrieverRequestKey struct{}
	loaderRequestKey    struct{}
	toolRequestKey      struct{}
	transformRequestKey struct{}
)

var (
	once       sync.Once
	openaiOnce sync.Once
	ollamaOnce sync.Once
	arkOnce    sync.Once
	qwenOnce   sync.Once
	claudeOnce sync.Once
)

const (
	OperationNameChat        = "chat"
	OperationNamePrompt      = "prompt"
	OperationNameEmbeddings  = "embeddings"
	OperationNameIndexer     = "indexer"
	OperationNameRetriever   = "retriever"
	OperationNameLoader      = "loader"
	OperationNameToolNode    = "tool_node"
	OperationNameExecuteTool = "execute_tool"
	OperationNameTransform   = "transform"
)

type einoRequest struct {
	operationName string
	input         map[string]interface{}
}

type einoResponse struct {
	operationName string
	output        map[string]interface{}
}

type einoLLMRequest struct {
	operationName    string
	modelName        string
	encodingFormats  []string
	frequencyPenalty float64
	presencePenalty  float64
	maxTokens        int64
	usageInputTokens int64
	stopSequences    []string
	temperature      float64
	topK             float64
	topP             float64
	serverAddress    string
	seed             int64
	input            []*schema.Message
}

type einoLLMResponse struct {
	responseFinishReasons []string
	responseModel         string
	usageOutputTokens     int64
	usageTotalTokens      int64
	responseID            string
	output                string
}

type ChatModelConfig struct {
	BaseURL          string
	PresencePenalty  float64
	Seed             int64
	FrequencyPenalty float64
	TopK             float64
}
