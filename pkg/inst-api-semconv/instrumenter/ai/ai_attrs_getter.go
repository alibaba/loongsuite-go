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

package ai

type CommonAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetAIOperationName(request REQUEST) string
	GetAISystem(request REQUEST) string
}

type LLMAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetAIRequestModel(request REQUEST) string
	GetAIRequestEncodingFormats(request REQUEST) []string
	GetAIRequestFrequencyPenalty(request REQUEST) float64
	GetAIRequestPresencePenalty(request REQUEST) float64
	GetAIResponseFinishReasons(request REQUEST, response RESPONSE) []string
	GetAIResponseModel(request REQUEST, response RESPONSE) string
	GetAIRequestMaxTokens(request REQUEST) int64
	GetAIUsageInputTokens(request REQUEST) int64
	GetAIUsageOutputTokens(request REQUEST, response RESPONSE) int64
	GetAIRequestStopSequences(request REQUEST) []string
	GetAIRequestTemperature(request REQUEST) float64
	GetAIRequestTopK(request REQUEST) float64
	GetAIRequestTopP(request REQUEST) float64
	GetAIResponseID(request REQUEST, response RESPONSE) string
	GetAIServerAddress(request REQUEST) string
	GetAIRequestSeed(request REQUEST) int64
	GetAIInput(request REQUEST) string
	GetAIOutput(response RESPONSE) string
}
