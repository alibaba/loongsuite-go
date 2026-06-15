package ai

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"testing"
)

type testRequest struct {
	System    string
	Operation string
}
type testResponse struct {
}
type commonRequest struct{}

func (commonRequest) GetAIOperationName(request testRequest) string {
	return request.Operation
}
func (commonRequest) GetAISystem(request testRequest) string {
	return request.System
}

type ollamaRequest struct{}

func (ollamaRequest) GetAIRequestModel(request testRequest) string {
	return "deepseek:17b"
}
func (ollamaRequest) GetAIRequestEncodingFormats(request testRequest) []string {
	return nil
}
func (ollamaRequest) GetAIRequestFrequencyPenalty(request testRequest) float64 {
	return 1.0
}
func (ollamaRequest) GetAIRequestPresencePenalty(request testRequest) float64 {
	return 1.0
}
func (ollamaRequest) GetAIResponseFinishReasons(request testRequest, response testResponse) []string {
	return []string{"stop"}
}
func (ollamaRequest) GetAIResponseModel(request testRequest, response testResponse) string {
	return "deepseek:17b"
}
func (ollamaRequest) GetAIRequestMaxTokens(request testRequest) int64 {
	return 10
}
func (ollamaRequest) GetAIUsageInputTokens(request testRequest) int64 {
	return 10
}
func (ollamaRequest) GetAIUsageOutputTokens(request testRequest, response testResponse) int64 {
	return 10
}
func (ollamaRequest) GetAIRequestStopSequences(request testRequest) []string {
	return []string{"stop"}
}
func (ollamaRequest) GetAIRequestTemperature(request testRequest) float64 {
	return 1.0
}
func (ollamaRequest) GetAIRequestTopK(request testRequest) float64 {
	return 1.0
}
func (ollamaRequest) GetAIRequestTopP(request testRequest) float64 {
	return 1.0
}

func (ollamaRequest) GetAIInput(request testRequest) string {
	return ""
}

func (ollamaRequest) GetAIOutput(response testResponse) string {
	return ""
}

func (ollamaRequest) GetAIResponseID(request testRequest, response testResponse) string {
	return "chatcmpl-123"
}
func (ollamaRequest) GetAIServerAddress(request testRequest) string {
	return "127.0.0.1:1234"
}

func (ollamaRequest) GetAIRequestSeed(request testRequest) int64 {
	return 100
}

func TestCommonExtractorStart(t *testing.T) {
	Extractor := AICommonAttrsExtractor[testRequest, any, commonRequest]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = Extractor.OnStart(attrs, parentContext, testRequest{Operation: "llm", System: "langchain"})
	if len(attrs) == 0 {
		t.Fatal("attrs is empty")
	}
	if attrs[0].Key != semconv.GenAIOperationNameKey || attrs[0].Value.AsString() != "llm" {
		t.Fatalf("gen_ai.operation.name be llm")
	}
	if attrs[1].Key != semconv.GenAISystemKey || attrs[1].Value.AsString() != "langchain" {
		t.Fatalf("gen_ai.system should be langchain")
	}
}
func TestCommonExtractorEnd(t *testing.T) {
	dbExtractor := AICommonAttrsExtractor[testRequest, any, commonRequest]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = dbExtractor.OnEnd(attrs, parentContext, testRequest{Operation: "llm", System: "langchain"}, testResponse{}, errors.New("test-err"))
	assert.Equal(t, semconv.ErrorTypeKey, attrs[0].Key)
	assert.Equal(t, "test-err", attrs[0].Value.AsString())
}

func TestAILLMAttrsExtractorStart(t *testing.T) {
	LLMExtractor := AILLMAttrsExtractor[testRequest, testResponse, commonRequest, ollamaRequest]{
		Base:      AICommonAttrsExtractor[testRequest, testResponse, commonRequest]{},
		LLMGetter: ollamaRequest{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = LLMExtractor.OnStart(attrs, parentContext, testRequest{Operation: "llm", System: "langchain"})
	if len(attrs) == 0 {
		t.Fatal("attrs is empty")
	}

	attrMap := make(map[attribute.Key]attribute.Value, len(attrs))
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	requireAttrString := func(key attribute.Key, want string) {
		val, ok := attrMap[key]
		assert.True(t, ok, "expected attribute %q to be present", key)
		if ok {
			assert.Equal(t, want, val.AsString())
		}
	}

	requireAttrInt64 := func(key attribute.Key, want int64) {
		val, ok := attrMap[key]
		assert.True(t, ok, "expected attribute %q to be present", key)
		if ok {
			assert.Equal(t, want, val.AsInt64())
		}
	}

	requireAttrFloat := func(key attribute.Key, want float64) {
		val, ok := attrMap[key]
		assert.True(t, ok, "expected attribute %q to be present", key)
		if ok {
			assert.InDelta(t, want, val.AsFloat64(), 1e-9)
		}
	}

	requireAttrString(semconv.GenAIOperationNameKey, "llm")
	requireAttrString(semconv.GenAISystemKey, "langchain")
	requireAttrString(semconv.GenAIRequestModelKey, "deepseek:17b")
	requireAttrInt64(semconv.GenAIRequestMaxTokensKey, 10)
	requireAttrFloat(semconv.GenAIRequestFrequencyPenaltyKey, 1.0)
	requireAttrFloat(semconv.GenAIRequestPresencePenaltyKey, 1.0)
	if val, ok := attrMap[semconv.GenAIRequestStopSequencesKey]; assert.True(t, ok) && ok {
		assert.Equal(t, []string{"stop"}, val.AsStringSlice())
	}
	requireAttrFloat(semconv.GenAIRequestTemperatureKey, 1.0)
	requireAttrFloat(semconv.GenAIRequestTopKKey, 1.0)
	requireAttrFloat(semconv.GenAIRequestTopPKey, 1.0)
	//	requireAttrInt64(semconv.GenAIUsageInputTokensKey, 10)
	requireAttrInt64(semconv.GenAIRequestSeedKey, 100)
	requireAttrString(semconv.ServerAddressKey, "127.0.0.1:1234")

	_, hasEncodingFormats := attrMap[semconv.GenAIRequestEncodingFormatsKey]
	assert.False(t, hasEncodingFormats, "encoding_formats attribute should be omitted")
}
func TestAILLMAttrsExtractorEnd(t *testing.T) {
	LLMExtractor := AILLMAttrsExtractor[testRequest, testResponse, commonRequest, ollamaRequest]{
		Base:      AICommonAttrsExtractor[testRequest, testResponse, commonRequest]{},
		LLMGetter: ollamaRequest{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = LLMExtractor.OnEnd(attrs, parentContext, testRequest{Operation: "llm", System: "langchain"}, testResponse{}, nil)
	attrMap := make(map[attribute.Key]attribute.Value, len(attrs))
	for _, attr := range attrs {
		attrMap[attr.Key] = attr.Value
	}

	requireValue := func(key attribute.Key) attribute.Value {
		val, ok := attrMap[key]
		assert.True(t, ok, "expected attribute %q to be present", key)
		return val
	}

	assert.Equal(t, []string{"stop"}, requireValue(semconv.GenAIResponseFinishReasonsKey).AsStringSlice())
	assert.Equal(t, "deepseek:17b", requireValue(semconv.GenAIResponseModelKey).AsString())
	assert.Equal(t, int64(10), requireValue(semconv.GenAIUsageInputTokensKey).AsInt64())
	assert.Equal(t, int64(10), requireValue(semconv.GenAIUsageOutputTokensKey).AsInt64())
	assert.Equal(t, "chatcmpl-123", requireValue(semconv.GenAIResponseIDKey).AsString())

	countResponseID := 0
	for _, attr := range attrs {
		if attr.Key == semconv.GenAIResponseIDKey {
			countResponseID++
		}
	}
	assert.Equal(t, 1, countResponseID, "response id attribute should appear exactly once")
}
