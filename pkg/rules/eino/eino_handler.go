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

package eino

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	callbacksutils "github.com/cloudwego/eino/utils/callbacks"
)

var (
	einoLLMInstrument    = BuildEinoLLMInstrumenter()
	einoCommonInstrument = BuildEinoCommonInstrumenter()
)

func einoModelCallHandler(config ChatModelConfig) *callbacksutils.ModelCallbackHandler {
	return &callbacksutils.ModelCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			request := einoLLMRequest{
				operationName:    OperationNameChat,
				serverAddress:    config.BaseURL,
				frequencyPenalty: config.FrequencyPenalty,
				presencePenalty:  config.PresencePenalty,
				seed:             config.Seed,
				topK:             config.TopK,
			}
			if input != nil {
				if input.Messages != nil && len(input.Messages) > 0 {
					request.input = input.Messages
				}
				if input.Config != nil {
					request.modelName = input.Config.Model
					request.maxTokens = int64(input.Config.MaxTokens)
					request.temperature = float64(input.Config.Temperature)
					request.stopSequences = input.Config.Stop
					request.topP = float64(input.Config.TopP)
				}
			}
			ctx = einoLLMInstrument.Start(ctx, request)
			return context.WithValue(ctx, llmRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			response := einoLLMResponse{}
			if output != nil {
				if output.TokenUsage != nil {
					response.usageOutputTokens = int64(output.TokenUsage.CompletionTokens)
					request.usageInputTokens = int64(output.TokenUsage.PromptTokens)
				}
				if output.Message != nil && output.Message.ResponseMeta != nil {
					response.responseFinishReasons = []string{output.Message.ResponseMeta.FinishReason}
				}
				if output.Config != nil {
					response.responseModel = output.Config.Model
				}
				if output.Message != nil {
					response.output = output.Message.Content
				}
			}
			einoLLMInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			response := einoLLMResponse{}
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, runInfo, string(debug.Stack()))
					}
					output.Close()
				}()
				firstTokenTime := time.Now()
				var outs []*model.CallbackOutput
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, runInfo)
					}
					outs = append(outs, chunk)
				}

				var usage *model.TokenUsage
				var mas []*schema.Message
				for _, out := range outs {
					if out == nil {
						continue
					}
					if out.TokenUsage != nil {
						usage = out.TokenUsage
					}
					if out.Message != nil {
						mas = append(mas, out.Message)
					}
				}
				if len(mas) != 0 {
					message, err := schema.ConcatMessages(mas)
					if err == nil {
						response.responseFinishReasons = []string{message.ResponseMeta.FinishReason}
						response.output = message.Content
					}
				}
				if usage != nil {
					response.usageOutputTokens = int64(usage.CompletionTokens)
					response.usageTotalTokens = int64(usage.TotalTokens)
					request.usageInputTokens = int64(usage.PromptTokens)
				}

				response.responseModel = request.modelName
				ctx = context.WithValue(ctx, ai.TimeToFirstTokenKey{}, firstTokenTime)
				einoLLMInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			response := einoLLMResponse{}
			einoLLMInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoPromptCallbackHandler() *callbacksutils.PromptCallbackHandler {
	return &callbacksutils.PromptCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *prompt.CallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNamePrompt,
				input:         make(map[string]any),
			}
			promptInput, err := sonic.MarshalString(input)
			if err == nil {
				request.input["prompt"] = promptInput
			}
			return startCommonInstrumentation(ctx, promptRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *prompt.CallbackOutput) context.Context {
			request := ctx.Value(promptRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNamePrompt,
				output:        make(map[string]any),
			}
			promptOutput, err := sonic.MarshalString(output)
			if err == nil {
				response.output["prompt"] = promptOutput
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(promptRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNamePrompt}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoEmbeddingCallbackHandler() *callbacksutils.EmbeddingCallbackHandler {
	return &callbacksutils.EmbeddingCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *embedding.CallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameEmbeddings,
				input:         make(map[string]any),
			}
			if input.Config != nil {
				request.input["encoding_format"] = input.Config.EncodingFormat
				request.input["model"] = input.Config.Model
				request.input["texts"] = strings.Join(input.Texts, ",")
			}
			return startCommonInstrumentation(ctx, embeddingRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *embedding.CallbackOutput) context.Context {
			request := ctx.Value(embeddingRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameEmbeddings,
				output:        make(map[string]any),
			}
			if output.Config != nil {
				response.output["encoding_format"] = output.Config.EncodingFormat
				response.output["model"] = output.Config.Model
				response.output["embeddings"] = output.Embeddings
				if output.TokenUsage != nil {
					response.output["prompt_tokens"] = output.TokenUsage.PromptTokens
					response.output["completion_tokens"] = output.TokenUsage.CompletionTokens
					response.output["total_tokens"] = output.TokenUsage.TotalTokens
				}
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(embeddingRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameEmbeddings}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoIndexerCallbackHandler() *callbacksutils.IndexerCallbackHandler {
	return &callbacksutils.IndexerCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *indexer.CallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameIndexer,
				input:         make(map[string]any),
			}
			docs, err := sonic.MarshalString(input)
			if err == nil {
				request.input["docs"] = docs
			}
			return startCommonInstrumentation(ctx, indexerRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *indexer.CallbackOutput) context.Context {
			request := ctx.Value(indexerRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameIndexer,
				output:        make(map[string]any),
			}
			if output != nil {
				response.output["doc_ids"] = output.IDs
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(indexerRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameIndexer}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoRetrieverCallbackHandler() *callbacksutils.RetrieverCallbackHandler {
	return &callbacksutils.RetrieverCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *retriever.CallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameRetriever,
				input:         make(map[string]any),
			}
			if input != nil {
				request.input["query"] = input.Query
				request.input["top_k"] = input.TopK
				request.input["filter"] = input.Filter
				if input.ScoreThreshold != nil {
					request.input["score_threshold"] = *input.ScoreThreshold
				}
			}
			return startCommonInstrumentation(ctx, retrieverRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *retriever.CallbackOutput) context.Context {
			request := ctx.Value(retrieverRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameRetriever,
				output:        make(map[string]any),
			}
			docs, err := sonic.MarshalString(output)
			if err == nil {
				response.output["docs"] = docs
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(retrieverRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameRetriever}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoLoaderCallbackHandler() *callbacksutils.LoaderCallbackHandler {
	return &callbacksutils.LoaderCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *document.LoaderCallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameLoader,
				input:         make(map[string]any),
			}
			if input != nil {
				request.input["source"] = input.Source.URI
			}
			return startCommonInstrumentation(ctx, loaderRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *document.LoaderCallbackOutput) context.Context {
			request := ctx.Value(loaderRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameLoader,
				output:        make(map[string]any),
			}
			loadOutput, err := sonic.MarshalString(output)
			if err == nil {
				response.output["docs"] = loadOutput
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(loaderRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameLoader}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoToolCallbackHandler() *callbacksutils.ToolCallbackHandler {
	return &callbacksutils.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameExecuteTool,
				input:         make(map[string]any),
			}
			if input != nil {
				request.input["arguments"] = input.ArgumentsInJSON
			}
			return startCommonInstrumentation(ctx, toolRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameExecuteTool,
				output:        make(map[string]any),
			}
			if output != nil {
				response.output["response"] = output.Response
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*tool.CallbackOutput]) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameExecuteTool,
				output:        make(map[string]any),
			}
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, info, string(debug.Stack()))
					}
					output.Close()
				}()
				var outs []*tool.CallbackOutput
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, info)
					}
					outs = append(outs, chunk)
				}
				toolResp := ""
				for _, out := range outs {
					if out == nil {
						continue
					}
					if out.Response != "" {
						toolResp += out.Response
					}
				}
				response.output["response"] = toolResp
				einoCommonInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameExecuteTool}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoToolsNodeCallbackHandler() *callbacksutils.ToolsNodeCallbackHandlers {
	return &callbacksutils.ToolsNodeCallbackHandlers{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *schema.Message) context.Context {
			request := einoRequest{
				operationName: OperationNameToolNode,
				input:         make(map[string]any),
			}
			if input != nil {
				request.input["role"] = input.Role
				if input.Content != "" {
					request.input["content"] = input.Content
				} else if len(input.MultiContent) > 0 {
					request.input["content"] = input.MultiContent
				}
				request.input["tool_call_id"] = input.ToolCallID
				request.input["tool_name"] = input.ToolName
				request.input["reasoning_content"] = input.ReasoningContent
			}
			return startCommonInstrumentation(ctx, toolRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output []*schema.Message) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameToolNode,
				output:        make(map[string]any),
			}
			for i, msg := range output {
				response.output[fmt.Sprintf("%d.role", i)] = msg.Role
				if msg.Content != "" {
					request.input[fmt.Sprintf("%d.content", i)] = msg.Content
				} else if len(msg.MultiContent) > 0 {
					request.input[fmt.Sprintf("%d.content", i)] = msg.MultiContent
				}
				response.output[fmt.Sprintf("%d.tool_call_id", i)] = msg.ToolCallID
				response.output[fmt.Sprintf("%d.tool_name", i)] = msg.ToolName
				response.output[fmt.Sprintf("%d.reasoning_content", i)] = msg.ReasoningContent
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[[]*schema.Message]) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameToolNode,
				output:        make(map[string]any),
			}
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, info, string(debug.Stack()))
					}
					output.Close()
				}()
				var outs []*schema.Message
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, info)
					}
					outs = append(outs, chunk...)
				}
				for i, msg := range outs {
					response.output[fmt.Sprintf("%d.role", i)] = msg.Role
					response.output[fmt.Sprintf("%d.content", i)] = msg.Content
					response.output[fmt.Sprintf("%d.tool_call_id", i)] = msg.ToolCallID
					response.output[fmt.Sprintf("%d.tool_name", i)] = msg.ToolName
					response.output[fmt.Sprintf("%d.reasoning_content", i)] = msg.ReasoningContent
				}
				einoCommonInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameToolNode}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoTransformCallbackHandler() *callbacksutils.TransformerCallbackHandler {
	return &callbacksutils.TransformerCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *document.TransformerCallbackInput) context.Context {
			request := einoRequest{
				operationName: OperationNameTransform,
				input:         make(map[string]any),
			}
			transformInput, err := sonic.MarshalString(input)
			if err == nil {
				request.input["docs"] = transformInput
			}
			return startCommonInstrumentation(ctx, transformRequestKey{}, request, einoCommonInstrument)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *document.TransformerCallbackOutput) context.Context {
			request := ctx.Value(transformRequestKey{}).(einoRequest)
			response := einoResponse{
				operationName: OperationNameTransform,
				output:        make(map[string]any),
			}
			transformOutput, err := sonic.MarshalString(output)
			if err == nil {
				response.output["docs"] = transformOutput
			}
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(transformRequestKey{}).(einoRequest)
			response := einoResponse{operationName: OperationNameTransform}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func startCommonInstrumentation(ctx context.Context, key any, req einoRequest, instrument instrumenter.Instrumenter[einoRequest, einoResponse]) context.Context {
	ctx = instrument.Start(ctx, req)
	return context.WithValue(ctx, key, req)
}

type ComposeHandler struct {
	operationName string
}

var _ callbacks.Handler = ComposeHandler{}

var _ callbacks.TimingChecker = ComposeHandler{}

func NewComposeHandler(operationName string) *ComposeHandler {
	return &ComposeHandler{
		operationName: operationName,
	}
}

func (c ComposeHandler) Needed(_ context.Context, runInfo *callbacks.RunInfo, timing callbacks.CallbackTiming) bool {
	return true
}

func (c ComposeHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	request := einoRequest{operationName: c.operationName}
	return einoCommonInstrument.Start(ctx, request)
}

func (c ComposeHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: c.operationName}
	einoCommonInstrument.End(ctx, request, response, nil)
	return ctx
}

func (c ComposeHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: c.operationName}
	einoCommonInstrument.End(ctx, request, response, err)
	return ctx
}

func (c ComposeHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	request := einoRequest{operationName: c.operationName}
	return einoCommonInstrument.Start(ctx, request)
}

func (c ComposeHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: c.operationName}
	einoCommonInstrument.End(ctx, request, response, nil)
	return ctx
}
