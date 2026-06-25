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

package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	openai "github.com/cohesion-org/deepseek-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var deepseekEnabled = os.Getenv("OTEL_INSTRUMENTATION_DEEPSEEK_ENABLED") != "false"

var deepseekTracer = otel.Tracer("github.com/cohesion-org/deepseek-go")

//go:linkname deepseekCreateChatCompletionOnEnter github.com/cohesion-org/deepseek-go.deepseekCreateChatCompletionOnEnter
func deepseekCreateChatCompletionOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request *openai.ChatCompletionRequest) {
	if !deepseekEnabled || request == nil {
		return
	}
	opts := append([]oteltrace.SpanStartOption{}, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	ctx, span := deepseekTracer.Start(ctx, "chat "+request.Model, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.model_name", request.Model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", request.Model),
		attribute.Int64("gen_ai.request.max_tokens", int64(request.MaxTokens)),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", "deepseek"),
	)

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var msgs []message
	for _, m := range request.Messages {
		msgs = append(msgs, message{Role: m.Role, Content: m.Content})
	}
	input, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(input)))
	}

	attrs = append(attrs, attribute.Int("gen_ai.max_tokens", request.MaxTokens))
	attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", float64(request.Temperature)),
		attribute.Bool("gen_ai.request.is_stream", false))
	attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", float64(request.TopP)),
		attribute.Float64("gen_ai.request.frequency_penalty", float64(request.FrequencyPenalty)),
		attribute.Float64("gen_ai.request.presence_penalty", float64(request.PresencePenalty)),
	)
	span.SetAttributes(attrs...)
	temp := make(map[string]interface{}, 3)
	temp["span"] = span
	temp["start"] = time.Now().UnixMilli()
	temp["model"] = request.Model
	call.SetData(temp)
	call.SetParam(1, ctx)
}

//go:linkname deepseekCreateChatCompletionOnExit github.com/cohesion-org/deepseek-go.deepseekCreateChatCompletionOnExit
func deepseekCreateChatCompletionOnExit(call api.CallContext, resp *openai.ChatCompletionResponse, err error) {
	if !deepseekEnabled || call.GetData() == nil {
		return
	}
	temp := call.GetData().(map[string]interface{})
	start := temp["start"].(int64)
	cost := time.Now().UnixMilli() - start
	span := temp["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	if resp == nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
		return
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	var reasons []string
	var msgs []interface{}
	for _, r := range resp.Choices {
		reasons = append(reasons, string(r.FinishReason))
		msgs = append(msgs, r.Message)
	}
	var spanAttrs []attribute.KeyValue
	out, err1 := json.Marshal(msgs)
	if err1 == nil {
		spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
	}
	spanAttrs = append(spanAttrs, attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
		attribute.Int64("gen_ai.response.time_to_first_token", cost*1000000),
		attribute.Int("gen_ai.usage.input_tokens", resp.Usage.PromptTokens),
		attribute.Int("gen_ai.usage.output_tokens", resp.Usage.CompletionTokens),
		attribute.Int("gen_ai.usage.total_tokens", resp.Usage.TotalTokens),
	)
	span.SetAttributes(spanAttrs...)
	span.End()
}

//go:linkname deepseekCreateChatCompletionStreamOnEnter github.com/cohesion-org/deepseek-go.deepseekCreateChatCompletionStreamOnEnter
func deepseekCreateChatCompletionStreamOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request *openai.StreamChatCompletionRequest) {
	if !deepseekEnabled || request == nil {
		return
	}
	var span oteltrace.Span
	opts := append([]oteltrace.SpanStartOption{}, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	ctx, span = deepseekTracer.Start(ctx, "chat "+request.Model, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.model_name", request.Model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", request.Model),
		attribute.Int64("gen_ai.request.max_tokens", int64(request.MaxTokens)),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", "deepseek"),
	)
	if request.ResponseFormat != nil {
		attrs = append(attrs, attribute.String("gen_ai.output.type", request.ResponseFormat.Type))
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var msgs []message
	for _, m := range request.Messages {
		msgs = append(msgs, message{Role: m.Role, Content: m.Content})
	}
	msgr, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgr)))
	}
	attrs = append(attrs, attribute.Int("gen_ai.max_tokens", request.MaxTokens))
	attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", float64(request.Temperature)))
	if len(request.Tools) > 0 {
		tools, _ := json.Marshal(request.Tools)
		attrs = append(attrs, attribute.String("gen_ai.tool.definitions", string(tools)))
	}
	if request.Stream {
		attrs = append(attrs, attribute.Bool("gen_ai.request.is_stream", true))
	}
	attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", float64(request.TopP)),
		attribute.Float64("gen_ai.request.frequency_penalty", float64(request.FrequencyPenalty)),
		attribute.Float64("gen_ai.request.presence_penalty", float64(request.PresencePenalty)),
	)
	span.SetAttributes(attrs...)
	temp := make(map[string]interface{}, 3)
	temp["span"] = span
	temp["start"] = time.Now().UnixMilli()
	temp["model"] = request.Model
	call.SetParam(1, ctx)
	call.SetData(temp)
}

//go:linkname deepseekCreateChatCompletionStreamOnExit github.com/cohesion-org/deepseek-go.deepseekCreateChatCompletionStreamOnExit
func deepseekCreateChatCompletionStreamOnExit(call api.CallContext, stream openai.ChatCompletionStream, err error) {
	if !deepseekEnabled || call.GetData() == nil {
		return
	}
	temp := call.GetData().(map[string]interface{})
	span := temp["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	model := temp["model"].(string)
	start := temp["start"].(int64)
	if x, ok := call.GetReturnVal(0).(openai.ChatCompletionStream); ok {
		st := &OTelChatStream{
			stream: x,
			span:   span,
			start:  start,
			model:  model,
		}
		call.SetReturnVal(0, st)
	} else {
		span.End()
	}
}

// OTelChatStream wraps a ChatCompletionStream to capture trace data.
type OTelChatStream struct {
	stream       openai.ChatCompletionStream
	span         oteltrace.Span
	inputTokens  int64
	outputTokens int64
	totalTokens  int64
	first        time.Time
	reasons      []string
	role         string
	content      string
	model        string
	start        int64
}

func (s *OTelChatStream) Recv() (*openai.StreamChatCompletionResponse, error) {
	c, err := s.stream.Recv()
	if err != nil || c == nil {
		if err != nil && !errors.Is(err, io.EOF) {
			s.span.SetStatus(codes.Error, err.Error())
		}
		var spanAttrs []attribute.KeyValue

		type message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		msgs := []message{{Role: s.role, Content: s.content}}
		out, err1 := json.Marshal(msgs)
		if err1 == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
		}

		var firstTokenCost int64
		if !s.first.IsZero() {
			firstTokenCost = s.first.UnixMilli() - s.start
		}
		spanAttrs = append(spanAttrs, attribute.Int64("gen_ai.usage.input_tokens", s.inputTokens),
			attribute.Int64("gen_ai.usage.output_tokens", s.outputTokens),
			attribute.Int64("gen_ai.usage.total_tokens", s.totalTokens),
			attribute.StringSlice("gen_ai.response.finish_reasons", s.reasons),
			attribute.Int64("gen_ai.response.time_to_first_token", firstTokenCost*1000000),
		)
		s.span.SetAttributes(spanAttrs...)
		s.span.End()
		return c, err
	}
	if c.Usage != nil {
		s.inputTokens = s.inputTokens + int64(c.Usage.PromptTokens)
		s.outputTokens = s.outputTokens + int64(c.Usage.CompletionTokens)
		s.totalTokens = s.totalTokens + int64(c.Usage.TotalTokens)
	}
	if s.first.IsZero() {
		s.first = time.Now()
	}
	for _, r := range c.Choices {
		if r.FinishReason != "" {
			s.reasons = append(s.reasons, string(r.FinishReason))
		}
		if r.Delta.Role != "" {
			s.role = r.Delta.Role
		}
		if r.Delta.Content != "" {
			s.content = s.content + r.Delta.Content
		}
	}
	return c, err
}

func (s *OTelChatStream) Close() error {
	return s.stream.Close()
}
