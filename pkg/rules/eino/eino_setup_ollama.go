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
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	utilscallbacks "github.com/cloudwego/eino/utils/callbacks"
)

//go:linkname ollamaGenerateOnEnter github.com/cloudwego/eino-ext/components/model/ollama.ollamaGenerateOnEnter
func ollamaGenerateOnEnter(call api.CallContext, cm *ollama.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}

//go:linkname ollamaStreamOnEnter github.com/cloudwego/eino-ext/components/model/ollama.ollamaStreamOnEnter
func ollamaStreamOnEnter(call api.CallContext, cm *ollama.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}
