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
	"os"
	"strings"

	openai "github.com/meguminnnnnnnnn/go-openai"
)

// openaiInnerEnabler controls whether OpenAI monitoring is enabled
type openaiInnerEnabler struct {
	enabled bool
}

func (o openaiInnerEnabler) Enable() bool {
	return o.enabled
}

var openaiEnabler = openaiInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_OPENAI_ENABLED") != "false"}

const (
	OperationNameChat           = "chat"
	OperationNameTextCompletion = "text_completion"
	OperationNameEmbeddings     = "embeddings"
)

// openaiRequest represents the monitoring data for an OpenAI API request
type openaiRequest struct {
	uid              string
	operationName    string
	modelName        string
	providerName     string
	frequencyPenalty float64
	presencePenalty  float64
	maxTokens        int64
	temperature      float64
	topP             float64
	seed             int64
	inputMessages    string
	isStream         bool
	inputTokens      int64
	stopSequences    []string
	serverAddress    string
}

// openaiResponse represents the monitoring data for an OpenAI API response
type openaiResponse struct {
	responseModel     string
	usageInputTokens  int64
	usageOutputTokens int64
	usageTotalTokens  int64
	responseID        string
	outputMessages    string
	choiceCount       int
	finishReasons     []string
}

// Provider detection mapping
type providerEntry struct {
	keyword  string
	provider string
}

var providerMapping = []providerEntry{
	{"openai.com", "openai"},
	{"azure.com", "azure"},
	{"anthropic.com", "anthropic"},
	{"dashscope.aliyuncs", "qwen"},
	{"volces.com", "ark"},
	{"ark.cn", "ark"},
	{"hunyuan", "tencent"},
	{"tencentcloudapi", "tencent"},
	{"googleapis.com", "google"},
	{"generativelanguage", "google"},
	{"deepseek.com", "deepseek"},
	{"moonshot", "moonshot"},
	{"zhipuai.cn", "zhipu"},
	{"bigmodel.cn", "zhipu"},
	{"baidu.com", "baidu"},
	{"minimax", "minimax"},
	{"siliconflow", "siliconflow"},
	{"together", "together"},
	{"mistral", "mistral"},
	{"groq.com", "groq"},
	{"ollama", "ollama"},
	{"localhost", "local"},
	{"127.0.0.1", "local"},
}

func getProviderName(client *openai.Client) string {
	if client == nil {
		return "openai"
	}
	rawURL := client.GetClientBaseURL()
	for _, entry := range providerMapping {
		if strings.Contains(rawURL, entry.keyword) {
			return entry.provider
		}
	}
	return "openai"
}
