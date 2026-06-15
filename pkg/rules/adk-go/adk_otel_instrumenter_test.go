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
	"testing"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func setupTracer() *tracetest.InMemoryExporter {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	mp := sdkmetric.NewMeterProvider()
	otel.SetMeterProvider(mp)
	ai.InitAIMetrics(mp.Meter("test"))

	return exporter
}

func getAttr(attrs []attribute.KeyValue, key string) attribute.Value {
	for _, a := range attrs {
		if string(a.Key) == key {
			return a.Value
		}
	}
	return attribute.Value{}
}

func TestLLMInstrumenter_ChatSpan(t *testing.T) {
	exporter := setupTracer()
	defer exporter.Reset()

	inst := BuildADKLLMInstrumenter()

	req := adkLLMRequest{
		operationName: OperationChat,
		modelName:     "gemini-2.5-flash",
		temperature:   0.7,
		topP:          0.9,
		topK:          40,
		maxTokens:     1024,
		stopSequences: []string{"END"},
		isStream:      false,
		inputMessages: `[{"role":"user","parts":[{"text":"hello"}]}]`,
	}

	ctx := inst.Start(context.Background(), req)

	resp := adkLLMResponse{
		responseModel:     "gemini-2.5-flash-001",
		usageInputTokens:  10,
		usageOutputTokens: 20,
		finishReasons:     []string{"STOP"},
		outputMessages:    `{"role":"model","parts":[{"text":"Hi!"}]}`,
	}

	inst.End(ctx, req, resp, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	if span.Name != OperationChat {
		t.Errorf("expected span name %q, got %q", OperationChat, span.Name)
	}
	if span.SpanKind != trace.SpanKindClient {
		t.Errorf("expected client span kind, got %v", span.SpanKind)
	}

	attrs := span.Attributes
	assertAttr(t, attrs, "gen_ai.operation.name", OperationChat)
	assertAttr(t, attrs, "gen_ai.system", SystemGoogleADK)
	assertAttr(t, attrs, "gen_ai.request.model", "gemini-2.5-flash")
	assertAttrFloat(t, attrs, "gen_ai.request.temperature", 0.7)
	assertAttrFloat(t, attrs, "gen_ai.request.top_p", 0.9)
	assertAttrFloat(t, attrs, "gen_ai.request.top_k", 40)
	assertAttrInt(t, attrs, "gen_ai.request.max_tokens", 1024)
	assertAttr(t, attrs, "gen_ai.response.model", "gemini-2.5-flash-001")
	assertAttrInt(t, attrs, "gen_ai.usage.output_tokens", 20)
	assertAttr(t, attrs, "gen_ai.span.kind", "generation")

	finishReasons := getAttr(attrs, "gen_ai.response.finish_reasons")
	if finishReasons.Type() == attribute.STRINGSLICE {
		reasons := finishReasons.AsStringSlice()
		if len(reasons) != 1 || reasons[0] != "STOP" {
			t.Errorf("expected finish_reasons [STOP], got %v", reasons)
		}
	}
}

func TestLLMInstrumenter_StreamingSpan(t *testing.T) {
	exporter := setupTracer()
	defer exporter.Reset()

	inst := BuildADKLLMInstrumenter()

	req := adkLLMRequest{
		operationName: OperationChat,
		modelName:     "gemini-2.5-flash",
		isStream:      true,
	}

	ctx := inst.Start(context.Background(), req)
	inst.End(ctx, req, adkLLMResponse{}, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes
	assertAttr(t, attrs, "gen_ai.system", SystemGoogleADK)
	assertAttr(t, attrs, "gen_ai.request.model", "gemini-2.5-flash")
}

func TestAgentInstrumenter_WorkflowSpan(t *testing.T) {
	exporter := setupTracer()
	defer exporter.Reset()

	inst := BuildADKAgentInstrumenter()

	req := adkAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		input: map[string]any{
			"session_id":   "sess-123",
			"user_id":      "user-456",
			"user_message": "What is the weather?",
		},
	}

	ctx := inst.Start(context.Background(), req)
	inst.End(ctx, req, nil, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name != OperationInvokeAgent {
		t.Errorf("expected span name %q, got %q", OperationInvokeAgent, span.Name)
	}

	attrs := span.Attributes
	assertAttr(t, attrs, "gen_ai.operation.name", OperationInvokeAgent)
	assertAttr(t, attrs, "gen_ai.system", SystemGoogleADK)
	assertAttr(t, attrs, "gen_ai.span.kind", "workflow")
	assertAttr(t, attrs, "gen_ai.other_input.session_id", "sess-123")
	assertAttr(t, attrs, "gen_ai.other_input.user_id", "user-456")
	assertAttr(t, attrs, "gen_ai.other_input.user_message", "What is the weather?")
}

func TestAgentInstrumenter_WithOutput(t *testing.T) {
	exporter := setupTracer()
	defer exporter.Reset()

	inst := BuildADKAgentInstrumenter()

	req := adkAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindAgent,
		input:         map[string]any{"query": "test"},
		output:        map[string]any{"result": "done"},
	}

	ctx := inst.Start(context.Background(), req)
	inst.End(ctx, req, nil, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes
	assertAttr(t, attrs, "gen_ai.span.kind", "agent")
	assertAttr(t, attrs, "gen_ai.other_input.query", "test")
	assertAttr(t, attrs, "gen_ai.other_output.result", "done")
}

func TestLLMRecorder(t *testing.T) {
	exporter := setupTracer()
	defer exporter.Reset()

	recorder := NewLLMRecorder()

	req := adkLLMRequest{
		operationName: OperationChat,
		modelName:     "gemini-2.0-pro",
		temperature:   1.0,
	}

	ctx := recorder.Start(context.Background(), req)
	resp := adkLLMResponse{
		responseModel:     "gemini-2.0-pro-001",
		usageInputTokens:  5,
		usageOutputTokens: 15,
		finishReasons:     []string{"STOP"},
	}
	recorder.End(ctx, req, resp, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	assertAttr(t, spans[0].Attributes, "gen_ai.request.model", "gemini-2.0-pro")
	assertAttr(t, spans[0].Attributes, "gen_ai.response.model", "gemini-2.0-pro-001")
}

func TestCommonGetters(t *testing.T) {
	getter := adkCommonRequest{}
	req := adkLLMRequest{operationName: "chat", modelName: "test-model"}

	if getter.GetAIOperationName(req) != "chat" {
		t.Errorf("expected operation name 'chat'")
	}
	if getter.GetAISystem(req) != SystemGoogleADK {
		t.Errorf("expected system %q", SystemGoogleADK)
	}
}

func TestLLMGetters(t *testing.T) {
	getter := adkLLMGetter{}
	req := adkLLMRequest{
		modelName:     "model-x",
		temperature:   0.5,
		topP:          0.8,
		topK:          50,
		maxTokens:     2048,
		stopSequences: []string{".", "!"},
		inputTokens:   100,
		inputMessages: "hello",
	}
	resp := adkLLMResponse{
		responseModel:     "model-x-v2",
		usageOutputTokens: 200,
		finishReasons:     []string{"STOP", "LENGTH"},
		outputMessages:    "world",
	}

	if getter.GetAIRequestModel(req) != "model-x" {
		t.Error("model mismatch")
	}
	if getter.GetAIRequestTemperature(req) != 0.5 {
		t.Error("temperature mismatch")
	}
	if getter.GetAIRequestTopP(req) != 0.8 {
		t.Error("topP mismatch")
	}
	if getter.GetAIRequestTopK(req) != 50 {
		t.Error("topK mismatch")
	}
	if getter.GetAIRequestMaxTokens(req) != 2048 {
		t.Error("maxTokens mismatch")
	}
	if len(getter.GetAIRequestStopSequences(req)) != 2 {
		t.Error("stopSequences mismatch")
	}
	if getter.GetAIUsageInputTokens(req) != 100 {
		t.Error("inputTokens mismatch")
	}
	if getter.GetAIInput(req) != "hello" {
		t.Error("input mismatch")
	}
	if getter.GetAIResponseModel(req, resp) != "model-x-v2" {
		t.Error("response model mismatch")
	}
	if getter.GetAIUsageOutputTokens(req, resp) != 200 {
		t.Error("outputTokens mismatch")
	}
	if len(getter.GetAIResponseFinishReasons(req, resp)) != 2 {
		t.Error("finishReasons mismatch")
	}
	if getter.GetAIOutput(resp) != "world" {
		t.Error("output mismatch")
	}
}

func TestAgentCommonGetters(t *testing.T) {
	getter := adkAgentCommonRequest{}
	req := adkAgentRequest{
		operationName: "invoke_agent",
		spanKind:      ai.GenAISpanKindAgent,
	}

	if getter.GetAIOperationName(req) != "invoke_agent" {
		t.Error("operation name mismatch")
	}
	if getter.GetAISystem(req) != SystemGoogleADK {
		t.Error("system mismatch")
	}
	if getter.GetGenAISpanKind(req) != ai.GenAISpanKindAgent {
		t.Error("span kind mismatch")
	}

	// Test empty span kind returns unknown
	emptyReq := adkAgentRequest{}
	if getter.GetGenAISpanKind(emptyReq) != ai.GenAISpanKindUnknown {
		t.Error("empty span kind should return unknown")
	}
}

func TestEnabler(t *testing.T) {
	e := adkGoInnerEnabler{enabled: true}
	if !e.Enable() {
		t.Error("expected enabled=true")
	}
	e = adkGoInnerEnabler{enabled: false}
	if e.Enable() {
		t.Error("expected enabled=false")
	}
}

func TestToAttrValue(t *testing.T) {
	tests := []struct {
		input    any
		expected attribute.Type
	}{
		{"hello", attribute.STRING},
		{42, attribute.INT64},
		{int64(100), attribute.INT64},
		{3.14, attribute.FLOAT64},
		{true, attribute.BOOL},
		{struct{ X int }{1}, attribute.STRING}, // fallback to string
	}
	for _, tt := range tests {
		val := toAttrValue(tt.input)
		if val.Type() != tt.expected {
			t.Errorf("toAttrValue(%v): expected type %v, got %v", tt.input, tt.expected, val.Type())
		}
	}
}

// --- Test helpers ---

func assertAttr(t *testing.T, attrs []attribute.KeyValue, key, expected string) {
	t.Helper()
	val := getAttr(attrs, key)
	if val.AsString() != expected {
		t.Errorf("attribute %q: expected %q, got %q", key, expected, val.AsString())
	}
}

func assertAttrFloat(t *testing.T, attrs []attribute.KeyValue, key string, expected float64) {
	t.Helper()
	val := getAttr(attrs, key)
	actual := val.AsFloat64()
	if actual < expected-0.01 || actual > expected+0.01 {
		t.Errorf("attribute %q: expected %f, got %f", key, expected, actual)
	}
}

func assertAttrInt(t *testing.T, attrs []attribute.KeyValue, key string, expected int64) {
	t.Helper()
	val := getAttr(attrs, key)
	if val.AsInt64() != expected {
		t.Errorf("attribute %q: expected %d, got %d", key, expected, val.AsInt64())
	}
}
