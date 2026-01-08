// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

type langChainRequest struct {
	operationName string
	system        string
	input         map[string]any
	output        map[string]any
}

type langChainLLMRequest struct {
	operationName    string
	moduleName       string
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
	input            string
}
type langChainLLMResponse struct {
	responseFinishReasons []string
	responseModel         string
	usageOutputTokens     int64
	responseID            string
	output                string
}
