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
	"context"
	"encoding/json"
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

//go:linkname newEmbeddingsOnEnter github.com/openai/openai-go/v2.newEmbeddingsOnEnter
func newEmbeddingsOnEnter(call api.CallContext, client *openai.Client, ctx context.Context, body openai.EmbeddingNewParams, opts ...option.RequestOption) {
	if !instrumentationEnabled {
		return
	}
	tracer := otel.GetTracerProvider().Tracer("loongsuite.instrumentation.openai")
	opts1 := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := tracer.Start(ctx, "embeddings "+body.Model, opts1...)
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.request.model", body.Model),
		attribute.String("gen_ai.operation.name", "embeddings"),
		attribute.String("gen_ai.provider.name", "openai"),
	)
	input, err := body.Input.MarshalJSON()
	if err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(input)))
	}
	span.SetAttributes(attrs...)
	temp := make(map[string]interface{}, 3)
	temp["span"] = span
	temp["start"] = time.Now().UnixMilli()
	call.SetData(temp)
	call.SetParam(1, ctx)
}

//go:linkname newEmbeddingsOnExit github.com/openai/openai-go/v2.newEmbeddingsOnExit
func newEmbeddingsOnExit(call api.CallContext, resp *openai.CreateEmbeddingResponse, err error) {
	if !instrumentationEnabled || call.GetData() == nil {
		return
	}
	temp := call.GetData().(map[string]interface{})
	start := temp["start"].(int64)
	cost := time.Now().UnixMilli() - start
	span := temp["span"].(trace.Span)
	if span == nil {
		return
	}
	var spanAttrs []attribute.KeyValue
	if resp != nil {
		out, err1 := json.Marshal(resp.Data)
		if err1 == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
		}
		spanAttrs = append(spanAttrs,
			attribute.Int64("gen_ai.response.time_to_first_token", cost*1000000),
			attribute.Int64("gen_ai.usage.input_tokens", resp.Usage.PromptTokens),
			attribute.Int64("gen_ai.usage.output_tokens", resp.Usage.TotalTokens),
			attribute.Int64("gen_ai.usage.total_tokens", resp.Usage.TotalTokens),
			attribute.String("gen_ai.response.model", string(resp.Model)),
		)
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.SetAttributes(spanAttrs...)
	span.End()
}
