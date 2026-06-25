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

package google_genai

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

func buildResponseAttrs(resp *genai.GenerateContentResponse) ([]attribute.KeyValue, int64, int64, int64) {
	var attrs []attribute.KeyValue
	var inputTokens, outputTokens, totalTokens int64

	if resp == nil {
		return attrs, 0, 0, 0
	}

	if resp.ResponseID != "" {
		attrs = append(attrs, attribute.String("gen_ai.response.id", resp.ResponseID))
	}
	if resp.ModelVersion != "" {
		attrs = append(attrs, attribute.String("gen_ai.response.model", resp.ModelVersion))
	}

	if resp.UsageMetadata != nil {
		inputTokens = int64(resp.UsageMetadata.PromptTokenCount)
		outputTokens = int64(resp.UsageMetadata.CandidatesTokenCount)
		totalTokens = int64(resp.UsageMetadata.TotalTokenCount)
		attrs = append(attrs,
			attribute.Int64("gen_ai.usage.input_tokens", inputTokens),
			attribute.Int64("gen_ai.usage.output_tokens", outputTokens),
			attribute.Int64("gen_ai.usage.total_tokens", totalTokens),
		)
	}

	var reasons []string
	var outputMsgs []Message
	for _, cand := range resp.Candidates {
		if cand == nil {
			continue
		}
		if cand.FinishReason != "" {
			reasons = append(reasons, string(cand.FinishReason))
		}
		if cand.Content != nil {
			var msg Message
			msg.Role = cand.Content.Role
			for _, p := range cand.Content.Parts {
				if p == nil {
					continue
				}
				var part Part
				if p.Text != "" {
					part.Content = p.Text
					part.Type = "text"
				}
				msg.Parts = append(msg.Parts, part)
			}
			outputMsgs = append(outputMsgs, msg)
		}
	}

	if len(reasons) > 0 {
		attrs = append(attrs, attribute.StringSlice("gen_ai.response.finish_reasons", reasons))
	}
	if outJSON, err := json.Marshal(outputMsgs); err == nil && len(outputMsgs) > 0 {
		attrs = append(attrs, attribute.String("gen_ai.output.messages", string(outJSON)))
	}

	return attrs, inputTokens, outputTokens, totalTokens
}

//go:linkname genaiGenerateContentOnEnter google.golang.org/genai.genaiGenerateContentOnEnter
func genaiGenerateContentOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "chat", opts...)

	reqAttrs := buildRequestAttrs(model, contents, config, false)
	span.SetAttributes(reqAttrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname genaiGenerateContentOnExit google.golang.org/genai.genaiGenerateContentOnExit
func genaiGenerateContentOnExit(call api.CallContext, resp *genai.GenerateContentResponse, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	start, _ := data["start"].(time.Time)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	respAttrs, _, _, _ := buildResponseAttrs(resp)
	cost := time.Since(start).Milliseconds()
	respAttrs = append(respAttrs, attribute.Int64("gen_ai.response.time_to_first_token", cost*1000000))
	span.SetAttributes(respAttrs...)
	span.SetStatus(codes.Ok, "")
}

//go:linkname genaiEmbedContentOnEnter google.golang.org/genai.genaiEmbedContentOnEnter
func genaiEmbedContentOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, contents []*genai.Content, config *genai.EmbedContentConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "embeddings", opts...)

	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.system", "google_genai"),
		attribute.String("embedding.model_name", model),
		attribute.String("gen_ai.span.kind", "EMBEDDING"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "embeddings"),
		attribute.String("gen_ai.provider.name", "google"),
	)
	if input, err := json.Marshal(contents); err == nil {
		attrs = append(attrs, attribute.String("embedding.embedding_input", string(input)))
	}
	span.SetAttributes(attrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname genaiEmbedContentOnExit google.golang.org/genai.genaiEmbedContentOnExit
func genaiEmbedContentOnExit(call api.CallContext, resp *genai.EmbedContentResponse, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	var spanAttrs []attribute.KeyValue
	if resp != nil {
		if output, marshalErr := json.Marshal(resp); marshalErr == nil {
			spanAttrs = append(spanAttrs, attribute.String("embedding.embedding_output", string(output)))
		}
	}
	span.SetAttributes(spanAttrs...)
	span.SetStatus(codes.Ok, "")
}

//go:linkname genaiGenerateImagesOnEnter google.golang.org/genai.genaiGenerateImagesOnEnter
func genaiGenerateImagesOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "generate_images", opts...)

	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.system", "google_genai"),
		attribute.String("gen_ai.model_name", model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "generate_content"),
		attribute.String("gen_ai.provider.name", "google"),
		attribute.String("gen_ai.request.prompt", prompt),
	)

	if config != nil {
		if config.NumberOfImages > 0 {
			attrs = append(attrs, attribute.Int("gen_ai.request.number_of_images", int(config.NumberOfImages)))
		}
		if config.OutputMIMEType != "" {
			attrs = append(attrs, attribute.String("gen_ai.request.output_mime_type", config.OutputMIMEType))
		}
		if config.AspectRatio != "" {
			attrs = append(attrs, attribute.String("gen_ai.request.aspect_ratio", config.AspectRatio))
		}
	}
	span.SetAttributes(attrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname genaiGenerateImagesOnExit google.golang.org/genai.genaiGenerateImagesOnExit
func genaiGenerateImagesOnExit(call api.CallContext, resp *genai.GenerateImagesResponse, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	var spanAttrs []attribute.KeyValue
	if resp != nil {
		spanAttrs = append(spanAttrs, attribute.Int("gen_ai.response.generated_images_count", len(resp.GeneratedImages)))
		if output, marshalErr := json.Marshal(resp.GeneratedImages); marshalErr == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.response.generated_images", string(output)))
		}
	}
	span.SetAttributes(spanAttrs...)
	span.SetStatus(codes.Ok, "")
}

//go:linkname generateVideosOnEnter google.golang.org/genai.generateVideosOnEnter
func generateVideosOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, prompt *string, image *genai.Image, video *genai.Video, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "generate_video", opts...)

	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.system", "google_genai"),
		attribute.String("gen_ai.model_name", model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "generate_content"),
		attribute.String("gen_ai.provider.name", "google"),
	)

	if image != nil {
		if imageData, err := json.Marshal(image); err == nil {
			attrs = append(attrs, attribute.String("gen_ai.request.image", string(imageData)))
		}
	}
	if video != nil {
		if videoData, err := json.Marshal(video); err == nil {
			attrs = append(attrs, attribute.String("gen_ai.request.video", string(videoData)))
		}
	}

	if config != nil {
		if config.Seed != nil {
			attrs = append(attrs, attribute.Int("gen_ai.request.seed", int(*config.Seed)))
		}
	}
	span.SetAttributes(attrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname generateVideosOnExit google.golang.org/genai.generateVideosOnExit
func generateVideosOnExit(call api.CallContext, r *genai.GenerateVideosOperation, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	var spanAttrs []attribute.KeyValue
	if r != nil && r.Response != nil {
		if output, marshalErr := json.Marshal(r.Response); marshalErr == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.response.generated_images", string(output)))
		}
	}
	span.SetAttributes(spanAttrs...)
	span.SetStatus(codes.Ok, "")
}

//go:linkname genaiUpscaleImageOnEnter google.golang.org/genai.genaiUpscaleImageOnEnter
func genaiUpscaleImageOnEnter(call api.CallContext, m genai.Models, ctx context.Context, model string, image *genai.Image, upscaleFactor string, config *genai.UpscaleImageConfig) {
	if !instrumentationEnabled {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)}
	ctx, span := otel.Tracer("google.golang.org/genai").Start(ctx, "upscale_image", opts...)

	var attrs []attribute.KeyValue
	attrs = append(attrs,
		attribute.String("gen_ai.system", "google_genai"),
		attribute.String("gen_ai.model_name", model),
		attribute.String("gen_ai.span.kind", "LLM"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "generate_content"),
		attribute.String("gen_ai.provider.name", "google"),
		attribute.String("gen_ai.request.upscale_factor", upscaleFactor),
	)

	if image != nil {
		if imageData, err := json.Marshal(image); err == nil {
			attrs = append(attrs, attribute.String("gen_ai.request.image", string(imageData)))
		}
	}

	if config != nil {
		if config.OutputMIMEType != "" {
			attrs = append(attrs, attribute.String("gen_ai.request.output_mime_type", config.OutputMIMEType))
		}
		if config.OutputGCSURI != "" {
			attrs = append(attrs, attribute.String("gen_ai.request.output_gcs_uri", config.OutputGCSURI))
		}
		if config.IncludeRAIReason {
			attrs = append(attrs, attribute.Bool("gen_ai.request.include_rai_reason", config.IncludeRAIReason))
		}
		if config.EnhanceInputImage {
			attrs = append(attrs, attribute.Bool("gen_ai.request.enhance_input_image", config.EnhanceInputImage))
		}
		if config.ImagePreservationFactor != nil {
			attrs = append(attrs, attribute.Float64("gen_ai.request.image_preservation_factor", float64(*config.ImagePreservationFactor)))
		}
	}
	span.SetAttributes(attrs...)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["model"] = model
	data["start"] = time.Now()
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname genaiUpscaleImageOnExit google.golang.org/genai.genaiUpscaleImageOnExit
func genaiUpscaleImageOnExit(call api.CallContext, resp *genai.UpscaleImageResponse, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	var spanAttrs []attribute.KeyValue
	if resp != nil {
		spanAttrs = append(spanAttrs, attribute.Int("gen_ai.response.generated_images_count", len(resp.GeneratedImages)))
		if output, marshalErr := json.Marshal(resp.GeneratedImages); marshalErr == nil {
			spanAttrs = append(spanAttrs, attribute.String("gen_ai.response.upscaled_images", string(output)))
		}
	}
	span.SetAttributes(spanAttrs...)
	span.SetStatus(codes.Ok, "")
}
