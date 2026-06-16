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

package adkgo

import (
	"context"
	"fmt"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// --- LLM span getters ---

type adkCommonRequest struct{}

func (adkCommonRequest) GetAIOperationName(request adkLLMRequest) string {
	return request.operationName
}

func (adkCommonRequest) GetAISystem(request adkLLMRequest) string {
	return SystemGoogleADK
}

type adkLLMGetter struct{}

func (adkLLMGetter) GetAIRequestModel(request adkLLMRequest) string {
	return request.modelName
}

func (adkLLMGetter) GetAIRequestEncodingFormats(request adkLLMRequest) []string {
	return nil
}

func (adkLLMGetter) GetAIRequestFrequencyPenalty(request adkLLMRequest) float64 {
	return 0
}

func (adkLLMGetter) GetAIRequestPresencePenalty(request adkLLMRequest) float64 {
	return 0
}

func (adkLLMGetter) GetAIResponseFinishReasons(request adkLLMRequest, response adkLLMResponse) []string {
	return response.finishReasons
}

func (adkLLMGetter) GetAIResponseModel(request adkLLMRequest, response adkLLMResponse) string {
	return response.responseModel
}

func (adkLLMGetter) GetAIRequestMaxTokens(request adkLLMRequest) int64 {
	return request.maxTokens
}

func (adkLLMGetter) GetAIUsageInputTokens(request adkLLMRequest) int64 {
	return request.inputTokens
}

func (adkLLMGetter) GetAIUsageOutputTokens(request adkLLMRequest, response adkLLMResponse) int64 {
	return response.usageOutputTokens
}

func (adkLLMGetter) GetAIRequestStopSequences(request adkLLMRequest) []string {
	return request.stopSequences
}

func (adkLLMGetter) GetAIRequestTemperature(request adkLLMRequest) float64 {
	return request.temperature
}

func (adkLLMGetter) GetAIRequestTopK(request adkLLMRequest) float64 {
	return request.topK
}

func (adkLLMGetter) GetAIRequestTopP(request adkLLMRequest) float64 {
	return request.topP
}

func (adkLLMGetter) GetAIResponseID(request adkLLMRequest, response adkLLMResponse) string {
	return ""
}

func (adkLLMGetter) GetAIServerAddress(request adkLLMRequest) string {
	return ""
}

func (adkLLMGetter) GetAIRequestSeed(request adkLLMRequest) int64 {
	return 0
}

func (adkLLMGetter) GetAIInput(request adkLLMRequest) string {
	return request.inputMessages
}

func (adkLLMGetter) GetAIOutput(response adkLLMResponse) string {
	return response.outputMessages
}

// ADKLLMAttrsExtractor adds the gen_ai.span.kind attribute for LLM spans.
type ADKLLMAttrsExtractor struct {
	Base ai.AILLMAttrsExtractor[adkLLMRequest, adkLLMResponse, adkCommonRequest, adkLLMGetter]
}

func (e ADKLLMAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request adkLLMRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, ai.GenAISpanKindGeneration.Attribute())
	return attributes, parentContext
}

func (e ADKLLMAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request adkLLMRequest, response adkLLMResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = e.Base.OnEnd(attributes, ctx, request, response, err)
	return attributes, ctx
}

// BuildADKLLMInstrumenter builds the instrumenter for ADK-Go LLM calls.
func BuildADKLLMInstrumenter() instrumenter.Instrumenter[adkLLMRequest, adkLLMResponse] {
	builder := instrumenter.Builder[adkLLMRequest, adkLLMResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[adkLLMRequest, adkLLMResponse]{
			Getter: adkCommonRequest{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[adkLLMRequest]{}).
		AddAttributesExtractor(&ADKLLMAttrsExtractor{
			Base: ai.AILLMAttrsExtractor[adkLLMRequest, adkLLMResponse, adkCommonRequest, adkLLMGetter]{
				Base: ai.AICommonAttrsExtractor[adkLLMRequest, adkLLMResponse, adkCommonRequest]{
					CommonGetter: adkCommonRequest{},
				},
				LLMGetter: adkLLMGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ADK_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(ai.AIClientMetrics("adk-go-client")).
		BuildInstrumenter()
}

// --- Agent span getters ---

type adkAgentCommonRequest struct{}

func (adkAgentCommonRequest) GetAIOperationName(request adkAgentRequest) string {
	return request.operationName
}

func (adkAgentCommonRequest) GetAISystem(request adkAgentRequest) string {
	return SystemGoogleADK
}

func (adkAgentCommonRequest) GetGenAISpanKind(request adkAgentRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindUnknown
	}
	return request.spanKind
}

// ADKAgentAttrsExtractor adds agent-specific attributes.
type ADKAgentAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[adkAgentRequest, any, adkAgentCommonRequest]
}

func (e ADKAgentAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request adkAgentRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	spanKind := adkAgentCommonRequest{}.GetGenAISpanKind(request)
	attributes = append(attributes, spanKind.Attribute())

	if request.input != nil {
		for k, v := range request.input {
			val := toAttrValue(v)
			if val.Type() > 0 {
				attributes = append(attributes, attribute.KeyValue{
					Key:   attribute.Key("gen_ai.other_input." + k),
					Value: val,
				})
			}
		}
	}
	return attributes, parentContext
}

func (e ADKAgentAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request adkAgentRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = e.Base.OnEnd(attributes, ctx, request, response, err)
	if request.output != nil {
		for k, v := range request.output {
			val := toAttrValue(v)
			if val.Type() > 0 {
				attributes = append(attributes, attribute.KeyValue{
					Key:   attribute.Key("gen_ai.other_output." + k),
					Value: val,
				})
			}
		}
	}
	return attributes, ctx
}

func toAttrValue(v any) attribute.Value {
	switch val := v.(type) {
	case string:
		return attribute.StringValue(val)
	case int:
		return attribute.IntValue(val)
	case int64:
		return attribute.Int64Value(val)
	case float64:
		return attribute.Float64Value(val)
	case bool:
		return attribute.BoolValue(val)
	default:
		return attribute.StringValue(fmt.Sprintf("%v", v))
	}
}

// BuildADKAgentInstrumenter builds the instrumenter for ADK-Go agent/workflow spans.
func BuildADKAgentInstrumenter() instrumenter.Instrumenter[adkAgentRequest, any] {
	builder := instrumenter.Builder[adkAgentRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[adkAgentRequest, any]{
			Getter: adkAgentCommonRequest{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[adkAgentRequest]{}).
		AddAttributesExtractor(&ADKAgentAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[adkAgentRequest, any, adkAgentCommonRequest]{
				CommonGetter: adkAgentCommonRequest{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ADK_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// LLMRecorder wraps the LLM instrumenter for convenient Start/End calls.
type LLMRecorder struct {
	instrumenter instrumenter.Instrumenter[adkLLMRequest, adkLLMResponse]
}

func NewLLMRecorder() *LLMRecorder {
	return &LLMRecorder{instrumenter: BuildADKLLMInstrumenter()}
}

func (r *LLMRecorder) Start(ctx context.Context, request adkLLMRequest) context.Context {
	return r.instrumenter.Start(ctx, request)
}

func (r *LLMRecorder) End(ctx context.Context, request adkLLMRequest, response adkLLMResponse, err error) {
	r.instrumenter.End(ctx, request, response, err)
}
