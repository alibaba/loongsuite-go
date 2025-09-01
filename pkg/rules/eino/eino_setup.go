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
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	utilscallbacks "github.com/cloudwego/eino/utils/callbacks"
)

//go:linkname newGraphOnEnter github.com/cloudwego/eino/compose.newGraphOnEnter
func newGraphOnEnter(call api.CallContext, cfg interface{}) {
	if !einoEnabler.Enable() {
		return
	}
	once.Do(func() {
		handler := utilscallbacks.NewHandlerHelper().
			Graph(NewComposeHandler("graph")).Chain(NewComposeHandler("chain")).
			Prompt(einoPromptCallbackHandler()).Transformer(einoTransformCallbackHandler()).
			Embedding(einoEmbeddingCallbackHandler()).Indexer(einoIndexerCallbackHandler()).
			Retriever(einoRetrieverCallbackHandler()).Loader(einoLoaderCallbackHandler()).
			Tool(einoToolCallbackHandler()).ToolsNode(einoToolsNodeCallbackHandler()).
			Lambda(NewComposeHandler("lambda")).
			Handler()
		callbacks.AppendGlobalHandlers(handler)
	})
}

//go:linkname openaiGenerateOnEnter github.com/cloudwego/eino-ext/components/model/openai.openaiGenerateOnEnter
func openaiGenerateOnEnter(call api.CallContext, cm *openai.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	openaiOnce.Do(func() {
		config := ChatModelConfig{}

		cli := reflect.ValueOf(*cm).FieldByName("cli")
		if cli.IsValid() && !cli.IsNil() {
			conf := cli.Elem().FieldByName("config")
			if conf.IsValid() && !conf.IsNil() {
				if conf.Elem().FieldByName("BaseURL").IsValid() {
					config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
				}
				if conf.Elem().FieldByName("PresencePenalty").IsValid() && !conf.Elem().FieldByName("PresencePenalty").IsNil() {
					config.PresencePenalty = conf.Elem().FieldByName("PresencePenalty").Elem().Float()
				}
				if conf.Elem().FieldByName("Seed").IsValid() && !conf.Elem().FieldByName("Seed").IsNil() {
					config.Seed = conf.Elem().FieldByName("Seed").Elem().Int()
				}
				if conf.Elem().FieldByName("FrequencyPenalty").IsValid() && !conf.Elem().FieldByName("FrequencyPenalty").IsNil() {
					config.FrequencyPenalty = conf.Elem().FieldByName("FrequencyPenalty").Elem().Float()
				}
			}
		}
		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "OpenAI Stream",
			Type:      "OpenAI",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)
		call.SetParam(1, ctx)
	})
}

//go:linkname openaiStreamOnEnter github.com/cloudwego/eino-ext/components/model/openai.openaiStreamOnEnter
func openaiStreamOnEnter(call api.CallContext, cm *openai.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	openaiOnce.Do(func() {
		config := ChatModelConfig{}

		cli := reflect.ValueOf(*cm).FieldByName("cli")
		if cli.IsValid() && !cli.IsNil() {
			conf := cli.Elem().FieldByName("config")
			if conf.IsValid() && !conf.IsNil() {
				if conf.Elem().FieldByName("BaseURL").IsValid() {
					config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
				}
				if conf.Elem().FieldByName("PresencePenalty").IsValid() && !conf.Elem().FieldByName("PresencePenalty").IsNil() {
					config.PresencePenalty = conf.Elem().FieldByName("PresencePenalty").Elem().Float()
				}
				if conf.Elem().FieldByName("Seed").IsValid() && !conf.Elem().FieldByName("Seed").IsNil() {
					config.Seed = conf.Elem().FieldByName("Seed").Elem().Int()
				}
				if conf.Elem().FieldByName("FrequencyPenalty").IsValid() && !conf.Elem().FieldByName("FrequencyPenalty").IsNil() {
					config.FrequencyPenalty = conf.Elem().FieldByName("FrequencyPenalty").Elem().Float()
				}
			}
		}
		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "OpenAI Stream",
			Type:      "OpenAI",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)
		call.SetParam(1, ctx)
	})
}

//go:linkname ollamaGenerateOnEnter github.com/cloudwego/eino-ext/components/model/ollama.ollamaGenerateOnEnter
func ollamaGenerateOnEnter(call api.CallContext, cm *ollama.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	ollamaOnce.Do(func() {
		config := ChatModelConfig{}

		conf := reflect.ValueOf(*cm).FieldByName("config")
		if conf.IsValid() && !conf.IsNil() {
			if conf.Elem().FieldByName("BaseURL").IsValid() {
				config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
			}
		}
		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Ollama Generate",
			Type:      "Ollama",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname ollamaStreamOnEnter github.com/cloudwego/eino-ext/components/model/ollama.ollamaStreamOnEnter
func ollamaStreamOnEnter(call api.CallContext, cm *ollama.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	ollamaOnce.Do(func() {
		config := ChatModelConfig{}

		conf := reflect.ValueOf(*cm).FieldByName("config")
		if conf.IsValid() && !conf.IsNil() {
			if conf.Elem().FieldByName("BaseURL").IsValid() {
				config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
			}
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Ollama Stream",
			Type:      "Ollama",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname arkGenerateOnEnter github.com/cloudwego/eino-ext/components/model/ark.arkGenerateOnEnter
func arkGenerateOnEnter(call api.CallContext, cm *ark.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	arkOnce.Do(func() {
		config := ChatModelConfig{}

		chatModel := reflect.ValueOf(*cm).FieldByName("chatModel")
		if chatModel.IsValid() && !chatModel.IsNil() {
			if chatModel.Elem().FieldByName("frequencyPenalty").IsValid() && !chatModel.Elem().FieldByName("frequencyPenalty").IsNil() {
				config.FrequencyPenalty = chatModel.Elem().FieldByName("frequencyPenalty").Elem().Float()
			}
			if chatModel.Elem().FieldByName("presencePenalty").IsValid() && !chatModel.Elem().FieldByName("presencePenalty").IsNil() {
				config.PresencePenalty = chatModel.Elem().FieldByName("presencePenalty").Elem().Float()
			}
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Ark Generate",
			Type:      "Ark",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname arkStreamOnEnter github.com/cloudwego/eino-ext/components/model/ark.arkStreamOnEnter
func arkStreamOnEnter(call api.CallContext, cm *ark.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	arkOnce.Do(func() {
		config := ChatModelConfig{}

		chatModel := reflect.ValueOf(*cm).FieldByName("chatModel")
		if chatModel.IsValid() && !chatModel.IsNil() {
			if chatModel.Elem().FieldByName("frequencyPenalty").IsValid() && !chatModel.Elem().FieldByName("frequencyPenalty").IsNil() {
				config.FrequencyPenalty = chatModel.Elem().FieldByName("frequencyPenalty").Elem().Float()
			}
			if chatModel.Elem().FieldByName("presencePenalty").IsValid() && !chatModel.Elem().FieldByName("presencePenalty").IsNil() {
				config.PresencePenalty = chatModel.Elem().FieldByName("presencePenalty").Elem().Float()
			}
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Ark Stream",
			Type:      "Ark",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname qwenGenerateOnEnter github.com/cloudwego/eino-ext/components/model/qwen.qwenGenerateOnEnter
func qwenGenerateOnEnter(call api.CallContext, cm *qwen.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	qwenOnce.Do(func() {
		config := ChatModelConfig{}

		cli := reflect.ValueOf(*cm).FieldByName("cli")
		if cli.IsValid() && !cli.IsNil() {
			conf := cli.Elem().FieldByName("config")
			if conf.IsValid() && !conf.IsNil() {
				if conf.Elem().FieldByName("BaseURL").IsValid() {
					config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
				}
				if conf.Elem().FieldByName("PresencePenalty").IsValid() && !conf.Elem().FieldByName("PresencePenalty").IsNil() {
					config.PresencePenalty = conf.Elem().FieldByName("PresencePenalty").Elem().Float()
				}
				if conf.Elem().FieldByName("Seed").IsValid() && !conf.Elem().FieldByName("Seed").IsNil() {
					config.Seed = conf.Elem().FieldByName("Seed").Elem().Int()
				}
				if conf.Elem().FieldByName("FrequencyPenalty").IsValid() && !conf.Elem().FieldByName("FrequencyPenalty").IsNil() {
					config.FrequencyPenalty = conf.Elem().FieldByName("FrequencyPenalty").Elem().Float()
				}
			}
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Qwen Generate",
			Type:      "Qwen",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname qwenStreamOnEnter github.com/cloudwego/eino-ext/components/model/qwen.qwenStreamOnEnter
func qwenStreamOnEnter(call api.CallContext, cm *qwen.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	qwenOnce.Do(func() {
		config := ChatModelConfig{}

		cli := reflect.ValueOf(*cm).FieldByName("cli")
		if cli.IsValid() && !cli.IsNil() {
			conf := cli.Elem().FieldByName("config")
			if conf.IsValid() && !conf.IsNil() {
				if conf.Elem().FieldByName("BaseURL").IsValid() {
					config.BaseURL = conf.Elem().FieldByName("BaseURL").String()
				}
				if conf.Elem().FieldByName("PresencePenalty").IsValid() && !conf.Elem().FieldByName("PresencePenalty").IsNil() {
					config.PresencePenalty = conf.Elem().FieldByName("PresencePenalty").Elem().Float()
				}
				if conf.Elem().FieldByName("Seed").IsValid() && !conf.Elem().FieldByName("Seed").IsNil() {
					config.Seed = conf.Elem().FieldByName("Seed").Elem().Int()
				}
				if conf.Elem().FieldByName("FrequencyPenalty").IsValid() && !conf.Elem().FieldByName("FrequencyPenalty").IsNil() {
					config.FrequencyPenalty = conf.Elem().FieldByName("FrequencyPenalty").Elem().Float()
				}
			}
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Qwen Stream",
			Type:      "Qwen",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname claudeGenerateOnEnter github.com/cloudwego/eino-ext/components/model/claude.claudeGenerateOnEnter
func claudeGenerateOnEnter(call api.CallContext, cm *claude.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	claudeOnce.Do(func() {
		config := ChatModelConfig{}

		topK := reflect.ValueOf(*cm).FieldByName("topK")
		if topK.IsValid() && !topK.IsNil() {
			config.TopK = topK.Elem().Float()
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Claude Generate",
			Type:      "Claude",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}

//go:linkname claudeStreamOnEnter github.com/cloudwego/eino-ext/components/model/claude.claudeStreamOnEnter
func claudeStreamOnEnter(call api.CallContext, cm *claude.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
	claudeOnce.Do(func() {
		config := ChatModelConfig{}

		topK := reflect.ValueOf(*cm).FieldByName("topK")
		if topK.IsValid() && !topK.IsNil() {
			config.TopK = topK.Elem().Float()
		}

		handler := utilscallbacks.NewHandlerHelper().ChatModel(einoModelCallHandler(config)).Handler()
		info := &callbacks.RunInfo{
			Name:      "Claude Stream",
			Type:      "Claude",
			Component: components.ComponentOfChatModel,
		}
		ctx = callbacks.InitCallbacks(ctx, info, handler)

		call.SetParam(1, ctx)
	})
}
