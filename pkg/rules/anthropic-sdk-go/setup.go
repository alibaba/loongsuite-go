// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package anthropic_sdk_go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname newClientOnEnter github.com/anthropics/anthropic-sdk-go.newClientOnEnter
func newClientOnEnter(call api.CallContext, opts ...option.RequestOption) {
	var options []option.RequestOption
	options = append(options, option.WithMiddleware(withTraceMiddleware()))
	if opts != nil {
		options = append(options, opts...)
	}
	call.SetParam(0, options)
}

func withTraceMiddleware() option.Middleware {
	return func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		if req == nil || req.Body == nil {
			return next(req)
		}
		if !strings.HasSuffix(req.URL.Path, "v1/messages") {
			return next(req)
		}

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return next(req)
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var request anthropic.MessageNewParams
		if err = json.Unmarshal(bodyBytes, &request); err != nil {
			return next(req)
		}

		inputMessages := ""
		if len(request.Messages) > 0 {
			if data, err := json.Marshal(request.Messages); err == nil {
				inputMessages = string(data)
			}
		}
		isStream := bytes.Contains(bodyBytes, []byte(`"stream":true`)) || bytes.Contains(bodyBytes, []byte(`"stream": true`))

		start := time.Now()
		span := chatSpanStart(req.Context(), request, inputMessages, isStream)

		resp, err := next(req)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.End()
			return resp, err
		}
		if resp == nil || resp.Body == nil {
			span.End()
			return resp, err
		}

		contentType := resp.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "text/event-stream") || isStream {
			resp.Body = &streamingReader{
				reader:  resp.Body,
				start:   start,
				span:    span,
				message: anthropic.Message{},
			}
			return resp, err
		}

		finalizeNonStream(resp, span)
		return resp, err
	}
}

func chatSpanStart(ctx context.Context, request anthropic.MessageNewParams, inputMessages string, isStream bool) oteltrace.Span {
	options := append([]oteltrace.SpanStartOption{}, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	spanName := "chat"
	if isStream {
		spanName = "chat stream"
	}
	_, span := otel.Tracer("github.com/anthropics/anthropic-sdk-go").Start(ctx, spanName, options...)

	var attrs []attribute.KeyValue
	model := string(request.Model)
	attrs = append(attrs,
		attribute.String("gen_ai.model_name", model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.Int64("gen_ai.request.max_tokens", request.MaxTokens),
		attribute.Int64("gen_ai.max_tokens", request.MaxTokens),
		attribute.Bool("gen_ai.request.is_stream", isStream),
	)
	if inputMessages != "" {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", inputMessages))
	}
	if request.Metadata.UserID.Valid() {
		attrs = append(attrs, attribute.String("gen_ai.user.id", request.Metadata.UserID.Value))
	}
	if request.Temperature.Valid() {
		attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", request.Temperature.Value))
	}
	if request.TopP.Valid() {
		attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", request.TopP.Value))
	}
	span.SetAttributes(attrs...)
	return span
}

func finalizeNonStream(resp *http.Response, span oteltrace.Span) {
	defer span.End()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var message anthropic.Message
	if err = json.Unmarshal(bodyBytes, &message); err != nil {
		return
	}
	spanAttrs := genAIResponseAttrs(message, time.Now())
	span.SetAttributes(spanAttrs...)
}

func genAIResponseAttrs(message anthropic.Message, first time.Time) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if output, err := json.Marshal(message.Content); err == nil {
		attrs = append(attrs, attribute.String("gen_ai.output.messages", string(output)))
	}
	if message.StopReason != "" {
		attrs = append(attrs, attribute.StringSlice("gen_ai.response.finish_reasons", []string{string(message.StopReason)}))
	}
	if !first.IsZero() {
		attrs = append(attrs, attribute.Int64("gen_ai.response.time_to_first_token", first.UnixMilli()*1000000))
	} else {
		attrs = append(attrs, attribute.Int64("gen_ai.response.time_to_first_token", time.Now().UnixMilli()*1000000))
	}
	usageInputTokens := totalInputTokens(message.Usage)
	usageOutputTokens := message.Usage.OutputTokens
	attrs = append(attrs,
		attribute.Int64("gen_ai.usage.input_tokens", usageInputTokens),
		attribute.Int64("gen_ai.usage.output_tokens", usageOutputTokens),
		attribute.Int64("gen_ai.usage.total_tokens", usageInputTokens+usageOutputTokens),
	)
	if message.ID != "" {
		attrs = append(attrs, attribute.String("gen_ai.response.id", message.ID))
	}
	return attrs
}

func totalInputTokens(usage anthropic.Usage) int64 {
	return usage.InputTokens + usage.CacheCreationInputTokens + usage.CacheReadInputTokens
}

type streamingReader struct {
	reader     io.ReadCloser
	teeReader  io.Reader
	logBuffer  *bytes.Buffer
	lineBuffer *bytes.Buffer
	start      time.Time
	first      time.Time
	span       oteltrace.Span
	message    anthropic.Message
	finished   bool
}

func (r *streamingReader) Read(p []byte) (int, error) {
	if r.teeReader == nil {
		r.logBuffer = &bytes.Buffer{}
		r.lineBuffer = &bytes.Buffer{}
		r.teeReader = io.TeeReader(r.reader, r.logBuffer)
	}

	n, err := r.teeReader.Read(p)
	if n > 0 {
		r.processSSELines()
	}
	if err != nil {
		r.finalize(err)
	}
	return n, err
}

func (r *streamingReader) Close() error {
	r.finalize(io.EOF)
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
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
	lines := bytes.Split(r.lineBuffer.Bytes(), []byte("\n"))
	var incompleteLine []byte
	for i, line := range lines {
		if i == len(lines)-1 {
			if len(line) > 0 {
				incompleteLine = append([]byte(nil), line...)
			}
			break
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 || !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}
		payload := bytes.TrimPrefix(line, []byte("data: "))
		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			continue
		}
		var event anthropic.MessageStreamEventUnion
		if err := json.Unmarshal(payload, &event); err == nil {
			if r.first.IsZero() && (event.Type == "content_block_delta" || event.Type == "message_delta") {
				r.first = time.Now()
			}
			_ = r.message.Accumulate(event)
		}
	}
	r.logBuffer.Reset()
	r.lineBuffer.Reset()
	if len(incompleteLine) > 0 {
		r.lineBuffer.Write(incompleteLine)
	}
}

func (r *streamingReader) finalize(streamErr error) {
	if r.finished {
		return
	}
	r.finished = true

	if streamErr != nil && !errors.Is(streamErr, io.EOF) {
		r.span.SetStatus(codes.Error, streamErr.Error())
	}
	spanAttrs := genAIResponseAttrs(r.message, r.first)
	r.span.SetAttributes(spanAttrs...)
	r.span.End()
}
