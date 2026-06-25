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

package newapi

import (
	"encoding/json"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	_ "unsafe"
)

func getChannelName(c *gin.Context) string {
	if name, ok := c.Get("channel_name"); ok {
		if s, ok2 := name.(string); ok2 {
			return s
		}
	}
	return ""
}

//go:linkname textHelperOnEnter github.com/QuantumNous/new-api/relay.textHelperOnEnter
func textHelperOnEnter(call api.CallContext, c *gin.Context, info *relaycommon.RelayInfo) {
	if !newAPIEnabler.Enable() {
		return
	}
	textReq, ok := info.Request.(*dto.GeneralOpenAIRequest)
	if !ok {
		return
	}
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	ctx, span := newAPITracer.Start(c.Request.Context(), "chat "+info.OriginModelName, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs, attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.user.id", strconv.Itoa(info.UserId)),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", getChannelName(c)),
	)
	if textReq.MaxTokens != nil {
		attrs = append(attrs, attribute.Int64("gen_ai.request.max_tokens", int64(*textReq.MaxTokens)))
	}
	if textReq.ResponseFormat != nil && textReq.ResponseFormat.Type != "" {
		attrs = append(attrs, attribute.String("gen_ai.output.type", textReq.ResponseFormat.Type))
	}

	var msgs []message
	for _, m := range textReq.Messages {
		var msg message
		if m.Role != "" {
			msg.Role = m.Role
		}
		var p part
		if m.Name != nil {
			p.Name = *m.Name
		}
		c1, err1 := json.Marshal(m.Content)
		if err1 == nil {
			p.Content = string(c1)
			p.Type = "text"
		}
		msg.Parts = append(msg.Parts, p)
		msgs = append(msgs, msg)
	}
	msgBytes, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgBytes)))
	}
	if textReq.Seed != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.seed", *textReq.Seed))
	}
	if textReq.Temperature != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", *textReq.Temperature))
	}
	if len(textReq.Tools) > 0 {
		tools, _ := json.Marshal(textReq.Tools)
		attrs = append(attrs, attribute.String("gen_ai.tool.definitions", string(tools)))
	}
	if textReq.TopP != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", *textReq.TopP))
	}
	if textReq.FrequencyPenalty != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.frequency_penalty", *textReq.FrequencyPenalty))
	}
	if textReq.PresencePenalty != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.presence_penalty", *textReq.PresencePenalty))
	}
	if textReq.N != nil {
		attrs = append(attrs, attribute.Int64("gen_ai.request.n", int64(*textReq.N)))
	}
	span.SetAttributes(attrs...)

	traceInfo := &streamTraceInfo{
		Span:  span,
		Model: textReq.Model,
	}
	c.Set(traceInfoCtxKey, traceInfo)

	data := make(map[string]interface{}, 2)
	data["info"] = info
	data["traceInfo"] = traceInfo
	c.Request = c.Request.WithContext(ctx)
	call.SetData(data)
	call.SetParam(0, c)
}

//go:linkname textHelperOnExit github.com/QuantumNous/new-api/relay.textHelperOnExit
func textHelperOnExit(call api.CallContext, newAPIError *types.NewAPIError) {
	if !newAPIEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	info, _ := data["info"].(*relaycommon.RelayInfo)
	traceInfo, _ := data["traceInfo"].(*streamTraceInfo)
	if info == nil || traceInfo == nil || traceInfo.Span == nil {
		return
	}
	span := traceInfo.Span
	if newAPIError != nil {
		span.SetStatus(codes.Error, newAPIError.Error())
	}
	var spanAttrs []attribute.KeyValue
	var msg message
	var reasons []string
	var p part
	spanAttrs = append(spanAttrs, attribute.String("gen_ai.session.id", info.RequestId),
		attribute.String("gen_ai.model_name", info.UpstreamModelName),
		attribute.String("gen_ai.request.model", info.OriginModelName),
		attribute.Int64("gen_ai.response.time_to_first_token", info.FirstResponseTime.Sub(info.StartTime).Nanoseconds()))
	if info.IsStream {
		for _, m := range traceInfo.Messages {
			var lastStreamResponse dto.ChatCompletionsStreamResponse
			if err := common.UnmarshalJsonStr(m, &lastStreamResponse); err != nil {
				continue
			}
			if lastStreamResponse.Choices != nil && len(lastStreamResponse.Choices) > 0 {
				for _, r := range lastStreamResponse.Choices {
					if r.FinishReason != nil {
						reasons = append(reasons, *r.FinishReason)
						msg.Reason = msg.Reason + *r.FinishReason
					}
					if r.Delta.Role != "" {
						msg.Role = r.Delta.Role
					}
					if r.Delta.Content != nil {
						p.Content = p.Content + *r.Delta.Content
						p.Type = "text"
					}
				}
			}
		}
	} else {
		for _, m := range traceInfo.Messages {
			var lastResponse []dto.OpenAITextResponseChoice
			if err := common.UnmarshalJsonStr(m, &lastResponse); err != nil {
				continue
			}
			for _, r := range lastResponse {
				if r.FinishReason != "" {
					reasons = append(reasons, r.FinishReason)
					msg.Reason = msg.Reason + r.FinishReason
				}
				msg.Role = r.Message.Role
				c1, err := json.Marshal(r.Message.Content)
				if err == nil {
					p.Content = p.Content + string(c1)
				}
			}
		}
	}

	spanAttrs = append(spanAttrs, attribute.Int("gen_ai.usage.input_tokens", traceInfo.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", traceInfo.OutputTokens),
		attribute.Int("gen_ai.usage.total_tokens", traceInfo.TotalTokens),
		attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
	)
	var msgs []message
	msg.Parts = append(msg.Parts, p)
	msgs = append(msgs, msg)
	out, err1 := json.Marshal(msgs)
	if err1 == nil {
		spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
	}
	span.SetAttributes(spanAttrs...)
	span.End()
}

//go:linkname audioHelperOnEnter github.com/QuantumNous/new-api/relay.audioHelperOnEnter
func audioHelperOnEnter(call api.CallContext, c *gin.Context, info *relaycommon.RelayInfo) {
	if !newAPIEnabler.Enable() {
		return
	}
	textReq, ok := info.Request.(*dto.AudioRequest)
	if !ok {
		return
	}
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	ctx, span := newAPITracer.Start(c.Request.Context(), "generate_content "+info.OriginModelName, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs, attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.user.id", strconv.Itoa(info.UserId)),
		attribute.String("gen_ai.operation.name", "generate_content"),
		attribute.String("gen_ai.provider.name", getChannelName(c)),
		attribute.String("gen_ai.input.messages", textReq.Input),
	)
	if textReq.ResponseFormat != "" {
		attrs = append(attrs, attribute.String("gen_ai.output.type", textReq.ResponseFormat))
	}
	span.SetAttributes(attrs...)

	traceInfo := &streamTraceInfo{
		Span:  span,
		Model: textReq.Model,
	}
	c.Set(traceInfoCtxKey, traceInfo)

	data := make(map[string]interface{}, 2)
	data["info"] = info
	data["traceInfo"] = traceInfo
	c.Request = c.Request.WithContext(ctx)
	call.SetData(data)
	call.SetParam(0, c)
}

//go:linkname audioHelperOnExit github.com/QuantumNous/new-api/relay.audioHelperOnExit
func audioHelperOnExit(call api.CallContext, newAPIError *types.NewAPIError) {
	if !newAPIEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	info, _ := data["info"].(*relaycommon.RelayInfo)
	traceInfo, _ := data["traceInfo"].(*streamTraceInfo)
	if info == nil || traceInfo == nil || traceInfo.Span == nil {
		return
	}
	span := traceInfo.Span
	if newAPIError != nil {
		span.SetStatus(codes.Error, newAPIError.Error())
	}
	var spanAttrs []attribute.KeyValue
	var reasons []string
	spanAttrs = append(spanAttrs, attribute.String("gen_ai.session.id", info.RequestId),
		attribute.String("gen_ai.model_name", info.UpstreamModelName),
		attribute.String("gen_ai.request.model", info.OriginModelName),
		attribute.Int64("gen_ai.response.time_to_first_token", info.FirstResponseTime.Sub(info.StartTime).Nanoseconds()))
	spanAttrs = append(spanAttrs, attribute.Int("gen_ai.usage.input_tokens", traceInfo.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", traceInfo.OutputTokens),
		attribute.Int("gen_ai.usage.total_tokens", traceInfo.TotalTokens),
		attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
	)
	span.SetAttributes(spanAttrs...)
	span.End()
}

//go:linkname claudeHelperOnEnter github.com/QuantumNous/new-api/relay.claudeHelperOnEnter
func claudeHelperOnEnter(call api.CallContext, c *gin.Context, info *relaycommon.RelayInfo) {
	if !newAPIEnabler.Enable() {
		return
	}
	textReq, ok := info.Request.(*dto.ClaudeRequest)
	if !ok {
		return
	}
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	ctx, span := newAPITracer.Start(c.Request.Context(), "chat "+info.OriginModelName, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs, attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.user.id", strconv.Itoa(info.UserId)),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", getChannelName(c)),
	)
	if textReq.MaxTokens != nil {
		attrs = append(attrs, attribute.Int64("gen_ai.request.max_tokens", int64(*textReq.MaxTokens)))
	}
	attrs = append(attrs, attribute.String("gen_ai.output.type", string(textReq.OutputFormat)))

	var msgs []message
	for _, m := range textReq.Messages {
		var msg message
		if m.Role != "" {
			msg.Role = m.Role
		}
		var p part
		c1, err1 := json.Marshal(m.Content)
		if err1 == nil {
			p.Content = string(c1)
			p.Type = "text"
		}
		msg.Parts = append(msg.Parts, p)
		msgs = append(msgs, msg)
	}
	msgBytes, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgBytes)))
	}
	if textReq.Temperature != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", *textReq.Temperature))
	}
	if textReq.TopP != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", *textReq.TopP))
	}
	span.SetAttributes(attrs...)

	traceInfo := &streamTraceInfo{
		Span:  span,
		Model: textReq.Model,
	}
	c.Set(traceInfoCtxKey, traceInfo)

	data := make(map[string]interface{}, 2)
	data["info"] = info
	data["traceInfo"] = traceInfo
	c.Request = c.Request.WithContext(ctx)
	call.SetData(data)
	call.SetParam(0, c)
}

//go:linkname claudeHelperOnExit github.com/QuantumNous/new-api/relay.claudeHelperOnExit
func claudeHelperOnExit(call api.CallContext, newAPIError *types.NewAPIError) {
	if !newAPIEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	info, _ := data["info"].(*relaycommon.RelayInfo)
	traceInfo, _ := data["traceInfo"].(*streamTraceInfo)
	if info == nil || traceInfo == nil || traceInfo.Span == nil {
		return
	}
	span := traceInfo.Span
	if newAPIError != nil {
		span.SetStatus(codes.Error, newAPIError.Error())
	}
	var spanAttrs []attribute.KeyValue
	var msg message
	var reasons []string
	var p part
	spanAttrs = append(spanAttrs, attribute.String("gen_ai.session.id", info.RequestId),
		attribute.String("gen_ai.model_name", info.UpstreamModelName),
		attribute.String("gen_ai.request.model", info.OriginModelName),
		attribute.Int64("gen_ai.response.time_to_first_token", info.FirstResponseTime.Sub(info.StartTime).Nanoseconds()))
	if info.IsStream {
		for _, m := range traceInfo.Messages {
			var lastStreamResponse dto.ChatCompletionsStreamResponse
			if err := common.UnmarshalJsonStr(m, &lastStreamResponse); err != nil {
				continue
			}
			if lastStreamResponse.Choices != nil && len(lastStreamResponse.Choices) > 0 {
				for _, r := range lastStreamResponse.Choices {
					if r.FinishReason != nil {
						reasons = append(reasons, *r.FinishReason)
						msg.Reason = msg.Reason + *r.FinishReason
					}
					if r.Delta.Role != "" {
						msg.Role = r.Delta.Role
					}
					if r.Delta.Content != nil {
						p.Content = p.Content + *r.Delta.Content
						p.Type = "text"
					}
				}
			}
		}
	} else {
		for _, m := range traceInfo.Messages {
			var lastResponse []dto.OpenAITextResponseChoice
			if err := common.UnmarshalJsonStr(m, &lastResponse); err != nil {
				continue
			}
			for _, r := range lastResponse {
				if r.FinishReason != "" {
					reasons = append(reasons, r.FinishReason)
					msg.Reason = msg.Reason + r.FinishReason
				}
				msg.Role = r.Message.Role
				c1, err := json.Marshal(r.Message.Content)
				if err == nil {
					p.Content = p.Content + string(c1)
				}
			}
		}
	}

	spanAttrs = append(spanAttrs, attribute.Int("gen_ai.usage.input_tokens", traceInfo.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", traceInfo.OutputTokens),
		attribute.Int("gen_ai.usage.total_tokens", traceInfo.TotalTokens),
		attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
	)
	var msgs []message
	msg.Parts = append(msg.Parts, p)
	msgs = append(msgs, msg)
	out, err1 := json.Marshal(msgs)
	if err1 == nil {
		spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
	}
	span.SetAttributes(spanAttrs...)
	span.End()
}

//go:linkname geminiHelperOnEnter github.com/QuantumNous/new-api/relay.geminiHelperOnEnter
func geminiHelperOnEnter(call api.CallContext, c *gin.Context, info *relaycommon.RelayInfo) {
	if !newAPIEnabler.Enable() {
		return
	}
	textReq, ok := info.Request.(*dto.GeminiChatRequest)
	if !ok {
		return
	}
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	ctx, span := newAPITracer.Start(c.Request.Context(), "chat "+info.OriginModelName, opts...)
	var attrs []attribute.KeyValue
	attrs = append(attrs, attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.user.id", strconv.Itoa(info.UserId)),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", getChannelName(c)),
	)
	if textReq.GenerationConfig.MaxOutputTokens != nil {
		attrs = append(attrs, attribute.Int64("gen_ai.request.max_tokens", int64(*textReq.GenerationConfig.MaxOutputTokens)))
	}

	var msgs []message
	for _, m := range textReq.Contents {
		var msg message
		if m.Role != "" {
			msg.Role = m.Role
		}
		for _, pp := range m.Parts {
			var pt part
			pt.Content = pp.Text
			pt.Type = "text"
			msg.Parts = append(msg.Parts, pt)
		}
		msgs = append(msgs, msg)
	}
	msgBytes, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgBytes)))
	}
	if textReq.GenerationConfig.Seed != nil {
		attrs = append(attrs, attribute.Int64("gen_ai.request.seed", *textReq.GenerationConfig.Seed))
	}
	if textReq.GenerationConfig.Temperature != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", *textReq.GenerationConfig.Temperature))
	}
	if len(textReq.Tools) > 0 {
		tools, _ := json.Marshal(textReq.Tools)
		attrs = append(attrs, attribute.String("gen_ai.tool.definitions", string(tools)))
	}
	if textReq.GenerationConfig.TopP != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", *textReq.GenerationConfig.TopP))
	}
	if textReq.GenerationConfig.FrequencyPenalty != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.frequency_penalty", float64(*textReq.GenerationConfig.FrequencyPenalty)))
	}
	if textReq.GenerationConfig.PresencePenalty != nil {
		attrs = append(attrs, attribute.Float64("gen_ai.request.presence_penalty", float64(*textReq.GenerationConfig.PresencePenalty)))
	}
	span.SetAttributes(attrs...)

	traceInfo := &streamTraceInfo{
		Span: span,
	}
	c.Set(traceInfoCtxKey, traceInfo)

	data := make(map[string]interface{}, 2)
	data["info"] = info
	data["traceInfo"] = traceInfo
	c.Request = c.Request.WithContext(ctx)
	call.SetData(data)
	call.SetParam(0, c)
}

//go:linkname geminiHelperOnExit github.com/QuantumNous/new-api/relay.geminiHelperOnExit
func geminiHelperOnExit(call api.CallContext, newAPIError *types.NewAPIError) {
	if !newAPIEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	info, _ := data["info"].(*relaycommon.RelayInfo)
	traceInfo, _ := data["traceInfo"].(*streamTraceInfo)
	if info == nil || traceInfo == nil || traceInfo.Span == nil {
		return
	}
	span := traceInfo.Span
	if newAPIError != nil {
		span.SetStatus(codes.Error, newAPIError.Error())
	}
	var spanAttrs []attribute.KeyValue
	var msg message
	var reasons []string
	var p part
	spanAttrs = append(spanAttrs, attribute.String("gen_ai.session.id", info.RequestId),
		attribute.String("gen_ai.model_name", info.UpstreamModelName),
		attribute.String("gen_ai.request.model", info.OriginModelName),
		attribute.Int64("gen_ai.response.time_to_first_token", info.FirstResponseTime.Sub(info.StartTime).Nanoseconds()))
	if info.IsStream {
		for _, m := range traceInfo.Messages {
			var lastStreamResponse dto.ChatCompletionsStreamResponse
			if err := common.UnmarshalJsonStr(m, &lastStreamResponse); err != nil {
				continue
			}
			if lastStreamResponse.Choices != nil && len(lastStreamResponse.Choices) > 0 {
				for _, r := range lastStreamResponse.Choices {
					if r.FinishReason != nil {
						reasons = append(reasons, *r.FinishReason)
						msg.Reason = msg.Reason + *r.FinishReason
					}
					if r.Delta.Role != "" {
						msg.Role = r.Delta.Role
					}
					if r.Delta.Content != nil {
						p.Content = p.Content + *r.Delta.Content
						p.Type = "text"
					}
				}
			}
		}
	} else {
		for _, m := range traceInfo.Messages {
			var lastResponse []dto.OpenAITextResponseChoice
			if err := common.UnmarshalJsonStr(m, &lastResponse); err != nil {
				continue
			}
			for _, r := range lastResponse {
				if r.FinishReason != "" {
					reasons = append(reasons, r.FinishReason)
					msg.Reason = msg.Reason + r.FinishReason
				}
				msg.Role = r.Message.Role
				c1, err := json.Marshal(r.Message.Content)
				if err == nil {
					p.Content = p.Content + string(c1)
				}
			}
		}
	}

	spanAttrs = append(spanAttrs, attribute.Int("gen_ai.usage.input_tokens", traceInfo.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", traceInfo.OutputTokens),
		attribute.Int("gen_ai.usage.total_tokens", traceInfo.TotalTokens),
		attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
	)
	var msgs []message
	msg.Parts = append(msg.Parts, p)
	msgs = append(msgs, msg)
	out, err1 := json.Marshal(msgs)
	if err1 == nil {
		spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
	}
	span.SetAttributes(spanAttrs...)
	span.End()
}
