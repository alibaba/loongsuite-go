package langchain

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"strings"
)

type aiLLMRequest struct {
}

var _ ai.LLMAttrsGetter[langChainLLMRequest, langChainLLMResponse] = aiLLMRequest{}
var _ ai.CommonAttrsGetter[langChainLLMRequest, any] = aiLLMRequest{}

func (aiLLMRequest) GetAIOperationName(request langChainLLMRequest) string {
	return request.operationName
}
func (aiLLMRequest) GetAISystem(request langChainLLMRequest) string {
	if request.moduleName == "" {
		return "langchain"
	}
	s := strings.Split(request.moduleName, ":")
	return s[0]
}
func (aiLLMRequest) GetAIRequestModel(request langChainLLMRequest) string {
	return request.moduleName
}
func (aiLLMRequest) GetAIRequestEncodingFormats(request langChainLLMRequest) []string {
	return request.encodingFormats
}
func (aiLLMRequest) GetAIRequestFrequencyPenalty(request langChainLLMRequest) float64 {
	return request.frequencyPenalty
}
func (aiLLMRequest) GetAIRequestPresencePenalty(request langChainLLMRequest) float64 {
	return request.presencePenalty
}
func (aiLLMRequest) GetAIRequestMaxTokens(request langChainLLMRequest) int64 {
	return request.maxTokens
}
func (aiLLMRequest) GetAIUsageInputTokens(request langChainLLMRequest) int64 {
	return request.usageInputTokens
}
func (aiLLMRequest) GetAIRequestStopSequences(request langChainLLMRequest) []string {
	return request.stopSequences
}
func (aiLLMRequest) GetAIRequestTemperature(request langChainLLMRequest) float64 {
	return request.temperature
}
func (aiLLMRequest) GetAIRequestTopK(request langChainLLMRequest) float64 {
	return request.topK
}
func (aiLLMRequest) GetAIRequestTopP(request langChainLLMRequest) float64 {
	return request.topP
}

func (aiLLMRequest) GetAIServerAddress(request langChainLLMRequest) string {
	return request.serverAddress
}
func (aiLLMRequest) GetAIRequestSeed(request langChainLLMRequest) int64 {
	return request.seed
}

func (aiLLMRequest) GetAIInput(request langChainLLMRequest) string {
	return request.input
}

func (aiLLMRequest) GetAIOutput(response langChainLLMResponse) string {
	return response.output
}

func (aiLLMRequest) GetAIUsageOutputTokens(request langChainLLMRequest, response langChainLLMResponse) int64 {
	return response.usageOutputTokens
}
func (aiLLMRequest) GetAIResponseID(request langChainLLMRequest, response langChainLLMResponse) string {
	return response.responseID
}
func (aiLLMRequest) GetAIResponseFinishReasons(request langChainLLMRequest, response langChainLLMResponse) []string {
	return response.responseFinishReasons
}
func (aiLLMRequest) GetAIResponseModel(request langChainLLMRequest, response langChainLLMResponse) string {
	return response.responseModel
}

var langChainLLMInstrument = BuildLangchainLLMOtelInstrumenter()

func BuildLangchainLLMOtelInstrumenter() instrumenter.Instrumenter[langChainLLMRequest, langChainLLMResponse] {
	builder := instrumenter.Builder[langChainLLMRequest, langChainLLMResponse]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[langChainLLMRequest, langChainLLMResponse]{Getter: aiLLMRequest{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[langChainLLMRequest]{}).
		AddAttributesExtractor(&ai.AILLMAttrsExtractor[langChainLLMRequest, langChainLLMResponse, aiLLMRequest, aiLLMRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.LANGCHAIN_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
