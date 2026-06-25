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
	"fmt"
	"io"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	openai "github.com/cohesion-org/deepseek-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname deepseekCreateFIMCompletionOnEnter github.com/cohesion-org/deepseek-go.deepseekCreateFIMCompletionOnEnter
func deepseekCreateFIMCompletionOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request *openai.FIMCompletionRequest) {
	if !deepseekEnabled || request == nil {
		return
	}
	opts := append([]oteltrace.SpanStartOption{}, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	ctx, span := deepseekTracer.Start(ctx, "text_completion "+request.Model, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.model_name", request.Model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", request.Model),
		attribute.Int64("gen_ai.request.max_tokens", int64(request.MaxTokens)),
		attribute.String("gen_ai.operation.name", "text_completion"),
		attribute.String("gen_ai.provider.name", "deepseek"),
	)

	// FIM specific attributes
	attrs = append(attrs, attribute.String("gen_ai.prompt", request.Prompt))
	if request.Suffix != "" {
		attrs = append(attrs, attribute.String("gen_ai.suffix", request.Suffix))
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

//go:linkname deepseekCreateFIMCompletionOnExit github.com/cohesion-org/deepseek-go.deepseekCreateFIMCompletionOnExit
func deepseekCreateFIMCompletionOnExit(call api.CallContext, resp *openai.FIMCompletionResponse, err error) {
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
	var texts []string
	for _, choice := range resp.Choices {
		reasons = append(reasons, choice.FinishReason)
		texts = append(texts, choice.Text)
	}

	var spanAttrs []attribute.KeyValue
	out, err1 := json.Marshal(texts)
	if err1 == nil {
		spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.text", string(out)))
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

//go:linkname deepseekCreateFIMStreamCompletionOnEnter github.com/cohesion-org/deepseek-go.deepseekCreateFIMStreamCompletionOnEnter
func deepseekCreateFIMStreamCompletionOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, request *openai.FIMStreamCompletionRequest) {
	if !deepseekEnabled || request == nil {
		return
	}
	var span oteltrace.Span
	opts := append([]oteltrace.SpanStartOption{}, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	ctx, span = deepseekTracer.Start(ctx, "text_completion "+request.Model, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.model_name", request.Model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", request.Model),
		attribute.Int64("gen_ai.request.max_tokens", int64(request.MaxTokens)),
		attribute.String("gen_ai.operation.name", "text_completion"),
		attribute.String("gen_ai.provider.name", "deepseek"),
	)

	// FIM specific attributes
	attrs = append(attrs, attribute.String("gen_ai.prompt", request.Prompt))
	if request.Suffix != "" {
		attrs = append(attrs, attribute.String("gen_ai.suffix", request.Suffix))
	}

	// Marshal prompt and suffix as input text for FIM
	inputText := map[string]string{
		"prompt": request.Prompt,
	}
	if request.Suffix != "" {
		inputText["suffix"] = request.Suffix
	}
	inputBytes, err := json.Marshal(inputText)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.text", string(inputBytes)))
	}

	attrs = append(attrs, attribute.Int("gen_ai.max_tokens", request.MaxTokens))
	attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", float64(request.Temperature)))
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

//go:linkname deepseekCreateFIMStreamCompletionOnExit github.com/cohesion-org/deepseek-go.deepseekCreateFIMStreamCompletionOnExit
func deepseekCreateFIMStreamCompletionOnExit(call api.CallContext, stream openai.FIMChatCompletionStream, err error) {
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
	if x, ok := call.GetReturnVal(0).(openai.FIMChatCompletionStream); ok {
		st := &OTelFIMStream{
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

// OTelFIMStream wraps a FIMChatCompletionStream to capture trace data.
type OTelFIMStream struct {
	stream       openai.FIMChatCompletionStream
	span         oteltrace.Span
	inputTokens  int64
	outputTokens int64
	totalTokens  int64
	first        time.Time
	reasons      []string
	text         string // accumulated text output
	model        string
	start        int64
}

func (s *OTelFIMStream) FIMRecv() (*openai.FIMStreamCompletionResponse, error) {
	c, err := s.stream.FIMRecv()
	if err != nil || c == nil {
		if err != nil && !errors.Is(err, io.EOF) {
			s.span.SetStatus(codes.Error, err.Error())
		}
		var spanAttrs []attribute.KeyValue

		// Marshal the accumulated text as output
		out, err1 := json.Marshal([]string{s.text})
		if err1 == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.text", string(out)))
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

	// Process usage if available in response
	if c.Usage != nil {
		usage := c.Usage
		if usage.PromptTokens > 0 {
			s.inputTokens += int64(usage.PromptTokens)
		}
		if usage.CompletionTokens > 0 {
			s.outputTokens += int64(usage.CompletionTokens)
		}
		if usage.TotalTokens > 0 {
			s.totalTokens += int64(usage.TotalTokens)
		}
	}

	// Record first token time
	if s.first.IsZero() {
		s.first = time.Now()
	}

	// Process choices from response
	for _, choice := range c.Choices {
		if choice.FinishReason != nil {
			var reasonStr string
			switch v := choice.FinishReason.(type) {
			case string:
				reasonStr = v
			default:
				reasonStr = fmt.Sprintf("%v", v)
			}
			if reasonStr != "" {
				s.reasons = append(s.reasons, reasonStr)
			}
		}

		// Accumulate the generated text
		s.text += choice.Text
	}

	return c, err
}

func (s *OTelFIMStream) FIMClose() error {
	return s.stream.FIMClose()
}
