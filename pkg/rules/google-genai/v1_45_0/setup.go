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

package v1_45_0

import (
	"context"
	"encoding/json"
	"os"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genai"
)

var instrumentationEnabled = os.Getenv("OTEL_INSTRUMENTATION_GOOGLE_GENAI_ENABLED") != "false"

// Message represents a conversation entry for JSON serialization.
type Message struct {
	Role   string `json:"role"`
	Parts  []Part `json:"parts"`
	Reason string `json:"finish_reason,omitempty"`
}

// Part represents part content within a message.
type Part struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

// spanContextKey is used to store span in context for streaming.
type spanContextKey struct{}

func buildRequestAttrs(model string, contents []*genai.Content, config *genai.GenerateContentConfig, isStream bool) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.system", "google_genai"),
		attribute.String("gen_ai.model_name", model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", "google"),
		attribute.Bool("gen_ai.request.is_stream", isStream),
	)

	if config != nil {
		if config.Temperature != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.temperature", float64(*config.Temperature)))
		}
		if config.TopP != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.top_p", float64(*config.TopP)))
		}
		if config.TopK != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.top_k", float64(*config.TopK)))
		}
		if config.MaxOutputTokens > 0 {
			attrs = append(attrs, attribute.Int64("gen_ai.request.max_tokens", int64(config.MaxOutputTokens)))
		}
		if config.FrequencyPenalty != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.frequency_penalty", float64(*config.FrequencyPenalty)))
		}
		if config.PresencePenalty != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.presence_penalty", float64(*config.PresencePenalty)))
		}
		if config.Seed != nil {
			attrs = append(attrs, attribute.Int64("gen_ai.request.seed", int64(*config.Seed)))
		}
		if len(config.StopSequences) > 0 {
			attrs = append(attrs, attribute.StringSlice("gen_ai.request.stop_sequences", config.StopSequences))
		}
	}

	var msgs []Message
	for _, c := range contents {
		if c == nil {
			continue
		}
		var msg Message
		msg.Role = c.Role
		for _, p := range c.Parts {
			if p == nil {
				continue
			}
			var part Part
			if p.Text != "" {
				part.Content = p.Text
				part.Type = "text"
			} else if p.FunctionCall != nil {
				fc, _ := json.Marshal(p.FunctionCall)
				part.Content = string(fc)
				part.Type = "function_call"
			} else if p.FunctionResponse != nil {
				fr, _ := json.Marshal(p.FunctionResponse)
				part.Content = string(fr)
				part.Type = "function_response"
			}
			msg.Parts = append(msg.Parts, part)
		}
		msgs = append(msgs, msg)
	}
	if msgJSON, err := json.Marshal(msgs); err == nil {
		attrs = append(attrs, attribute.String("gen_ai.input.messages", string(msgJSON)))
	}

	return attrs
}

//go:linkname genaiGenerateContentStreamOnEnter google.golang.org/genai.genaiGenerateContentStreamOnEnter
func genaiGenerateContentStreamOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "chat", opts...)
	respChan := make(chan genai.SpanResp, 10)
	ctx = context.WithValue(ctx, "arms_resp_chan", respChan)
	ctx = context.WithValue(ctx, spanContextKey{}, span)

	var first bool
	var reasons []string
	var msg Message
	var inputTokens int32
	var outputTokens int32
	var totalTokens int32
	start := time.Now().UnixMilli()
	var spanAttrs []attribute.KeyValue
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(respChan)
				return
			case resp, ok1 := <-respChan:
				if resp.Err != nil || resp.Resp == nil || !ok1 {
					if resp.Err != nil {
						span.SetStatus(codes.Error, resp.Err.Error())
					} else {
						span.SetStatus(codes.Ok, "")
					}

					var msgs []Message
					msgs = append(msgs, msg)
					out, err1 := json.Marshal(msgs)
					if err1 == nil {
						spanAttrs = append(spanAttrs, attribute.String("gen_ai.output.messages", string(out)))
					}

					spanAttrs = append(spanAttrs, attribute.Int("gen_ai.usage.input_tokens", int(inputTokens)),
						attribute.Int("gen_ai.usage.output_tokens", int(outputTokens)),
						attribute.Int("gen_ai.usage.total_tokens", int(totalTokens)),
						attribute.StringSlice("gen_ai.response.finish_reasons", reasons),
					)
					span.SetAttributes(spanAttrs...)
					span.End()
					return
				}
				if !first {
					first = true
					firstTokenCost := time.Now().UnixMilli() - start
					spanAttrs = append(spanAttrs, attribute.String("gen_ai.response.id", resp.Resp.ResponseID),
						attribute.Int64("gen_ai.response.time_to_first_token", firstTokenCost*1000000))
				}
				for _, r := range resp.Resp.Candidates {
					if r.FinishReason != "" {
						reasons = append(reasons, string(r.FinishReason))
						msg.Reason = msg.Reason + string(r.FinishReason)
					}
					if r.Content != nil && r.Content.Role != "" {
						msg.Role = r.Content.Role
					}
					var part Part
					if r.Content != nil {
						for _, r1 := range r.Content.Parts {
							part.Content = part.Content + r1.Text
						}
						part.Type = "text"
					}
					msg.Parts = append(msg.Parts, part)
				}
				if resp.Resp.UsageMetadata != nil {
					inputTokens = inputTokens + resp.Resp.UsageMetadata.PromptTokenCount
					outputTokens = outputTokens + resp.Resp.UsageMetadata.TotalTokenCount - resp.Resp.UsageMetadata.PromptTokenCount
					totalTokens = totalTokens + resp.Resp.UsageMetadata.TotalTokenCount
				}
			}
		}
	}()
	reqAttrs := buildRequestAttrs(model, contents, config, true)
	span.SetAttributes(reqAttrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}
