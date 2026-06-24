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

package openai_go_v2

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var instrumentationEnabled = os.Getenv("OTEL_INSTRUMENTATION_OPENAI_ENABLED") != "false"

var providerMapping = map[string]string{
	"openai.com":         "openai",
	"azure.com":          "azure",
	"anthropic.com":      "anthropic",
	"dashscope.aliyuncs": "qwen",
	"volces.com":         "ark",
	"ark.cn":             "ark",
	"hunyuan":            "tencent",
	"tencentcloudapi":    "tencent",
	"googleapis.com":     "google",
	"generativelanguage": "google",
	"deepseek.com":       "deepseek",
	"moonshot":           "moonshot",
	"zhipuai.cn":         "zhipu",
	"bigmodel.cn":        "zhipu",
	"baidu.com":          "baidu",
	"minimax":            "minimax",
	"siliconflow":        "siliconflow",
	"together":           "together",
	"mistral":            "mistral",
	"groq.com":           "groq",
	"ollama":             "ollama",
	"localhost":          "local",
	"127.0.0.1":          "local",
}

func getProviderName(rawURL string) string {
	for keyword, provider := range providerMapping {
		if strings.Contains(rawURL, keyword) {
			return provider
		}
	}
	return "openai"
}

// message is a local type for serializing gen_ai input/output messages
type message struct {
	Role   string `json:"role,omitempty"`
	Parts  []part `json:"parts,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type part struct {
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
}

// content is used for JSON unmarshaling of message content
type content struct {
	Msg string `json:"content"`
}

//go:linkname newClientOnEnter github.com/openai/openai-go/v2.newClientOnEnter
func newClientOnEnter(call api.CallContext, opts ...option.RequestOption) {
	if !instrumentationEnabled {
		return
	}
	var optTemp []option.RequestOption
	optTemp = append(optTemp, option.WithMiddleware(WithTraceMiddleware()))
	if opts != nil {
		optTemp = append(optTemp, opts...)
	}
	opts = optTemp
	call.SetParam(0, opts)
}

func chatSpanStart(req *http.Request, request openai.ChatCompletionNewParams, provider string) trace.Span {
	tracer := otel.GetTracerProvider().Tracer("loongsuite.instrumentation.openai")
	opts1 := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	_, span := tracer.Start(req.Context(), "chat "+request.Model, opts1...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.request.model", request.Model),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", provider),
		attribute.Int64("gen_ai.request.max_tokens", request.MaxTokens.Value),
	)
	if request.ResponseFormat.GetType() != nil {
		attrs = append(attrs, attribute.String("gen_ai.output.type", *request.ResponseFormat.GetType()))
	}
	var msgs []message
	for _, m := range request.Messages {
		var msg message
		if m.GetRole() != nil {
			msg.Role = *m.GetRole()
		}
		var p part
		if m.GetName() != nil {
			p.Name = *m.GetName()
		}
		c, err1 := m.MarshalJSON()
		if err1 == nil {
			var mc content
			err2 := json.Unmarshal(c, &mc)
			if err2 == nil {
				p.Content = mc.Msg
				p.Type = "text"
			}
		}
		msg.Parts = append(msg.Parts, p)
		msgs = append(msgs, msg)
	}
	msgJSON, err := json.Marshal(msgs)
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgJSON)))
	}
	attrs = append(attrs,
		attribute.Int64("gen_ai.request.seed", request.Seed.Value),
		attribute.Float64("gen_ai.request.temperature", request.Temperature.Value),
		attribute.Float64("gen_ai.request.top_p", request.TopP.Value),
		attribute.Float64("gen_ai.request.frequency_penalty", request.FrequencyPenalty.Value),
		attribute.Float64("gen_ai.request.presence_penalty", request.PresencePenalty.Value),
	)
	if len(request.Tools) > 0 {
		tools, _ := json.Marshal(request.Tools)
		attrs = append(attrs, attribute.String("gen_ai.tool.definitions", string(tools)))
	}
	span.SetAttributes(attrs...)
	return span
}

func completionSpanStart(req *http.Request, request openai.CompletionNewParams, provider string) trace.Span {
	tracer := otel.GetTracerProvider().Tracer("loongsuite.instrumentation.openai")
	opts1 := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	_, span := tracer.Start(req.Context(), "text_completion "+string(request.Model), opts1...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.request.model", string(request.Model)),
		attribute.String("gen_ai.operation.name", "text_completion"),
		attribute.String("gen_ai.provider.name", provider),
		attribute.Int64("gen_ai.request.max_tokens", request.MaxTokens.Value),
	)

	// Handle prompt (can be string or array)
	var msgs []message
	var msg message
	promptData, err1 := request.Prompt.MarshalJSON()
	if err1 == nil {
		var promptValue interface{}
		if err2 := json.Unmarshal(promptData, &promptValue); err2 == nil {
			var promptStr string
			if str, ok := promptValue.(string); ok {
				promptStr = str
			} else if arr, ok := promptValue.([]interface{}); ok && len(arr) > 0 {
				if str, ok := arr[0].(string); ok {
					promptStr = str
				}
			}
			if promptStr != "" {
				var p part
				p.Content = promptStr
				p.Type = "text"
				msg.Parts = append(msg.Parts, p)
				msgs = append(msgs, msg)
				msgJSON, err3 := json.Marshal(msgs)
				if err3 == nil {
					attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgJSON)))
				}
			}
		}
	}
	attrs = append(attrs,
		attribute.Int64("gen_ai.request.seed", request.Seed.Value),
		attribute.Float64("gen_ai.request.temperature", request.Temperature.Value),
		attribute.Float64("gen_ai.request.top_p", request.TopP.Value),
		attribute.Float64("gen_ai.request.frequency_penalty", request.FrequencyPenalty.Value),
		attribute.Float64("gen_ai.request.presence_penalty", request.PresencePenalty.Value),
	)
	span.SetAttributes(attrs...)
	return span
}

func WithTraceMiddleware() option.Middleware {
	return func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		start := time.Now()
		if !instrumentationEnabled || req.Body == nil {
			return next(req)
		}
		path := req.URL.Path
		var model string
		var span trace.Span
		provider := getProviderName(req.URL.Host)
		if strings.HasSuffix(path, "chat/completions") {
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil {
				req.Body.Close()
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				var request openai.ChatCompletionNewParams
				if err1 := json.Unmarshal(bodyBytes, &request); err1 == nil {
					model = request.Model
					span = chatSpanStart(req, request, provider)
				} else {
					return next(req)
				}
			} else {
				return next(req)
			}

			resp, err := next(req)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return resp, err
			}
			contentType := resp.Header.Get("Content-Type")
			isStreaming := strings.HasPrefix(contentType, "text/event-stream")
			if isStreaming {
				span.SetAttributes(attribute.Bool("gen_ai.request.stream", true))
				wrappedReader := &streamingReader{
					reader: resp.Body,
					start:  start,
					span:   span,
					model:  model,
				}
				resp.Body = wrappedReader
			} else {
				defer span.End()
				bodyBytes1, err1 := io.ReadAll(resp.Body)
				if err1 == nil {
					resp.Body.Close()
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes1))
					var r openai.ChatCompletion
					if err2 := json.Unmarshal(bodyBytes1, &r); err2 == nil {
						cost := time.Since(start).Milliseconds()
						var spanAttrs []attribute.KeyValue
						var reasons []string
						var msgs []message
						for _, r1 := range r.Choices {
							if r1.FinishReason != "" {
								reasons = append(reasons, string(r1.FinishReason))
							}
							var msg message
							msg.Role = string(r1.Message.Role)
							var p part
							p.Content = r1.Message.Content
							msg.Parts = append(msg.Parts, p)
							msgs = append(msgs, msg)
						}
						output, err2 := json.Marshal(msgs)
						if err2 == nil {
							spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(output)))
						}
						spanAttrs = append(spanAttrs,
							attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
							attribute.Int64("gen_ai.response.time_to_first_token", cost*1000000),
							attribute.Int64("gen_ai.usage.input_tokens", r.Usage.PromptTokens),
							attribute.Int64("gen_ai.usage.output_tokens", r.Usage.CompletionTokens),
							attribute.Int64("gen_ai.usage.total_tokens", r.Usage.TotalTokens),
							attribute.String("gen_ai.response.id", r.ID),
							attribute.String("gen_ai.response.model", r.Model),
						)
						span.SetAttributes(spanAttrs...)
					}
				}
			}
			return resp, err
		} else if strings.HasSuffix(path, "completions") && !strings.HasSuffix(path, "chat/completions") {
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil {
				req.Body.Close()
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				var request openai.CompletionNewParams
				if err1 := json.Unmarshal(bodyBytes, &request); err1 == nil {
					model = string(request.Model)
					span = completionSpanStart(req, request, provider)
				} else {
					return next(req)
				}
			} else {
				return next(req)
			}

			resp, err := next(req)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return resp, err
			}
			contentType := resp.Header.Get("Content-Type")
			isStreaming := strings.HasPrefix(contentType, "text/event-stream")
			if isStreaming {
				span.SetAttributes(attribute.Bool("gen_ai.request.stream", true))
				wrappedReader := &completionStreamingReader{
					reader: resp.Body,
					start:  start,
					span:   span,
					model:  model,
				}
				resp.Body = wrappedReader
			} else {
				defer span.End()
				bodyBytes1, err1 := io.ReadAll(resp.Body)
				if err1 == nil {
					resp.Body.Close()
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes1))
					var r openai.Completion
					if err2 := json.Unmarshal(bodyBytes1, &r); err2 == nil {
						cost := time.Since(start).Milliseconds()
						var spanAttrs []attribute.KeyValue
						var reasons []string
						var msgs []message
						for _, r1 := range r.Choices {
							if r1.FinishReason != "" {
								reasons = append(reasons, string(r1.FinishReason))
							}
							var msg message
							var p part
							p.Content = r1.Text
							p.Type = "text"
							msg.Parts = append(msg.Parts, p)
							msgs = append(msgs, msg)
						}
						output, err2 := json.Marshal(msgs)
						if err2 == nil {
							spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(output)))
						}
						spanAttrs = append(spanAttrs,
							attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
							attribute.Int64("gen_ai.response.time_to_first_token", cost*1000000),
							attribute.Int64("gen_ai.usage.input_tokens", r.Usage.PromptTokens),
							attribute.Int64("gen_ai.usage.output_tokens", r.Usage.CompletionTokens),
							attribute.Int64("gen_ai.usage.total_tokens", r.Usage.TotalTokens),
							attribute.String("gen_ai.response.id", r.ID),
							attribute.String("gen_ai.response.model", r.Model),
						)
						span.SetAttributes(spanAttrs...)
					}
				}
			}
			return resp, err
		} else {
			return next(req)
		}
	}
}

// streamingReader wraps an io.ReadCloser to parse SSE stream for chat completions tracing
type streamingReader struct {
	reader       io.ReadCloser
	teeReader    io.Reader
	logBuffer    *bytes.Buffer
	lineBuffer   *bytes.Buffer
	start        time.Time
	inputTokens  int64
	outputTokens int64
	totalTokens  int64
	id           string
	reasons      []string
	msg          message
	part         part
	span         trace.Span
	first        time.Time
	model        string
}

func (r *streamingReader) Read(p []byte) (n int, err error) {
	if r.teeReader == nil {
		r.logBuffer = &bytes.Buffer{}
		r.lineBuffer = &bytes.Buffer{}
		r.teeReader = io.TeeReader(r.reader, r.logBuffer)
	}

	n, err = r.teeReader.Read(p)

	if n > 0 {
		r.processSSELines()
	} else {
		var firstTokenCost int64
		cost := time.Since(r.start).Milliseconds()
		if !r.first.IsZero() {
			firstTokenCost = r.first.Sub(r.start).Milliseconds()
		} else {
			firstTokenCost = cost
		}
		var spanAttrs []attribute.KeyValue
		var msgs []message
		r.msg.Parts = append(r.msg.Parts, r.part)
		msgs = append(msgs, r.msg)
		out, err1 := json.Marshal(msgs)
		if err1 == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
		}
		spanAttrs = append(spanAttrs,
			attribute.StringSlice("gen_ai.response.finish_reasons", r.reasons),
			attribute.Int64("gen_ai.response.time_to_first_token", firstTokenCost*1000000),
			attribute.Int64("gen_ai.usage.input_tokens", r.inputTokens),
			attribute.Int64("gen_ai.usage.output_tokens", r.outputTokens),
			attribute.Int64("gen_ai.usage.total_tokens", r.totalTokens),
			attribute.String("gen_ai.response.id", r.id),
		)
		r.span.SetAttributes(spanAttrs...)
		r.span.End()
	}

	return n, err
}

func (r *streamingReader) processSSELines() {
	if r.logBuffer == nil || r.logBuffer.Len() == 0 {
		return
	}

	data := r.logBuffer.Bytes()
	if len(data) == 0 {
		return
	}

	r.lineBuffer.Write(data)

	allData := r.lineBuffer.Bytes()
	lines := bytes.Split(allData, []byte("\n"))

	var incompleteLine []byte
	for i, line := range lines {
		if i == len(lines)-1 {
			if len(line) > 0 {
				incompleteLine = line
			}
			break
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		a, done := r.processSSELine(line)
		if !done && a != nil {
			var rc openai.ChatCompletionChunk
			err1 := json.Unmarshal(a, &rc)
			if err1 == nil {
				if r.first.IsZero() {
					r.first = time.Now()
				}
				r.inputTokens = r.inputTokens + rc.Usage.PromptTokens
				r.outputTokens = r.outputTokens + rc.Usage.CompletionTokens
				r.totalTokens = r.totalTokens + rc.Usage.TotalTokens
				r.id = rc.ID
				for _, rc1 := range rc.Choices {
					if rc1.FinishReason != "" {
						r.reasons = append(r.reasons, string(rc1.FinishReason))
						r.msg.Reason = r.msg.Reason + string(rc1.FinishReason)
					}
					if rc1.Delta.Role != "" {
						r.msg.Role = rc1.Delta.Role
					}
					if rc1.Delta.Content != "" {
						r.part.Content = r.part.Content + rc1.Delta.Content
						r.part.Type = "text"
					}
				}
			}
		}
	}

	r.logBuffer.Reset()
	r.lineBuffer.Reset()
	if len(incompleteLine) > 0 {
		r.lineBuffer.Write(incompleteLine)
	}
}

func (r *streamingReader) processSSELine(line []byte) ([]byte, bool) {
	if bytes.HasPrefix(line, []byte("data: ")) {
		payload := bytes.TrimPrefix(line, []byte("data: "))
		if bytes.Equal(payload, []byte("[DONE]")) {
			return payload, true
		}
		return payload, false
	}
	return nil, false
}

func (r *streamingReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}

// completionStreamingReader wraps an io.ReadCloser to parse SSE stream for text completions tracing
type completionStreamingReader struct {
	reader       io.ReadCloser
	teeReader    io.Reader
	logBuffer    *bytes.Buffer
	lineBuffer   *bytes.Buffer
	start        time.Time
	inputTokens  int64
	outputTokens int64
	totalTokens  int64
	id           string
	reasons      []string
	msg          message
	part         part
	span         trace.Span
	first        time.Time
	model        string
}

func (r *completionStreamingReader) Read(p []byte) (n int, err error) {
	if r.teeReader == nil {
		r.logBuffer = &bytes.Buffer{}
		r.lineBuffer = &bytes.Buffer{}
		r.teeReader = io.TeeReader(r.reader, r.logBuffer)
	}

	n, err = r.teeReader.Read(p)

	if n > 0 {
		r.processSSELines()
	} else {
		var firstTokenCost int64
		cost := time.Since(r.start).Milliseconds()
		if !r.first.IsZero() {
			firstTokenCost = r.first.Sub(r.start).Milliseconds()
		} else {
			firstTokenCost = cost
		}
		var spanAttrs []attribute.KeyValue
		var msgs []message
		r.msg.Parts = append(r.msg.Parts, r.part)
		msgs = append(msgs, r.msg)
		out, err1 := json.Marshal(msgs)
		if err1 == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
		}
		spanAttrs = append(spanAttrs,
			attribute.StringSlice("gen_ai.response.finish_reasons", r.reasons),
			attribute.Int64("gen_ai.response.time_to_first_token", firstTokenCost*1000000),
			attribute.Int64("gen_ai.usage.input_tokens", r.inputTokens),
			attribute.Int64("gen_ai.usage.output_tokens", r.outputTokens),
			attribute.Int64("gen_ai.usage.total_tokens", r.totalTokens),
			attribute.String("gen_ai.response.id", r.id),
		)
		r.span.SetAttributes(spanAttrs...)
		r.span.End()
	}

	return n, err
}

func (r *completionStreamingReader) processSSELines() {
	if r.logBuffer == nil || r.logBuffer.Len() == 0 {
		return
	}

	data := r.logBuffer.Bytes()
	if len(data) == 0 {
		return
	}

	r.lineBuffer.Write(data)

	allData := r.lineBuffer.Bytes()
	lines := bytes.Split(allData, []byte("\n"))

	var incompleteLine []byte
	for i, line := range lines {
		if i == len(lines)-1 {
			if len(line) > 0 {
				incompleteLine = line
			}
			break
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		a, done := r.processSSELine(line)
		if !done && a != nil {
			var rc openai.Completion
			err1 := json.Unmarshal(a, &rc)
			if err1 == nil {
				if r.first.IsZero() {
					r.first = time.Now()
				}
				r.inputTokens = r.inputTokens + rc.Usage.PromptTokens
				r.outputTokens = r.outputTokens + rc.Usage.CompletionTokens
				r.totalTokens = r.totalTokens + rc.Usage.TotalTokens
				r.id = rc.ID
				for _, rc1 := range rc.Choices {
					if rc1.FinishReason != "" {
						r.reasons = append(r.reasons, string(rc1.FinishReason))
						r.msg.Reason = r.msg.Reason + string(rc1.FinishReason)
					}
					if rc1.Text != "" {
						r.part.Content = r.part.Content + rc1.Text
						r.part.Type = "text"
					}
				}
			}
		}
	}

	r.logBuffer.Reset()
	r.lineBuffer.Reset()
	if len(incompleteLine) > 0 {
		r.lineBuffer.Write(incompleteLine)
	}
}

func (r *completionStreamingReader) processSSELine(line []byte) ([]byte, bool) {
	if bytes.HasPrefix(line, []byte("data: ")) {
		payload := bytes.TrimPrefix(line, []byte("data: "))
		if bytes.Equal(payload, []byte("[DONE]")) {
			return payload, true
		}
		return payload, false
	}
	return nil, false
}

func (r *completionStreamingReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}
