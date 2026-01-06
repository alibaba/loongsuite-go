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
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	utilscallbacks "github.com/cloudwego/eino/utils/callbacks"
)

//go:linkname claudeGenerateOnEnter github.com/cloudwego/eino-ext/components/model/claude.claudeGenerateOnEnter
func claudeGenerateOnEnter(call api.CallContext, cm *claude.ChatModel, ctx context.Context, input []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}

//go:linkname claudeStreamOnEnter github.com/cloudwego/eino-ext/components/model/claude.claudeStreamOnEnter
func claudeStreamOnEnter(call api.CallContext, cm *claude.ChatModel, ctx context.Context, in []*schema.Message, opts ...model.Option) {
	if !einoEnabler.Enable() {
		return
	}
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
}
