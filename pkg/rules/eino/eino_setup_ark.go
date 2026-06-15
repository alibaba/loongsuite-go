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
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	utilscallbacks "github.com/cloudwego/eino/utils/callbacks"
)

//go:linkname arkGenerateOnEnter github.com/cloudwego/eino-ext/components/model/ark.arkGenerateOnEnter
func arkGenerateOnEnter(call api.CallContext, cm *ark.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}

//go:linkname arkStreamOnEnter github.com/cloudwego/eino-ext/components/model/ark.arkStreamOnEnter
func arkStreamOnEnter(call api.CallContext, cm *ark.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}
