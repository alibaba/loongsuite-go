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

package mcp

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type aiCommonRequest struct {
}

func (aiCommonRequest) GetAIOperationName(request mcpRequest) string {
	return request.operationName
}
func (aiCommonRequest) GetAISystem(request mcpRequest) string {
	return request.system
}

type LExperimentalAttributeExtractor struct {
	Base ai.AICommonAttrsExtractor[mcpRequest, any, aiCommonRequest]
}

func (l LExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request mcpRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = l.Base.OnStart(attributes, parentContext, request)
	var val attribute.Value
	if request.methodType == string(mcp.MethodToolsCall) {
		attributes = append(attributes, attribute.KeyValue{
			Key:   "gen_ai.tool.name",
			Value: attribute.StringValue(request.methodName),
		}, attribute.KeyValue{
			Key:   "gen_ai.tool.call.id",
			Value: attribute.StringValue(request.CallId),
		})
	}
	if request.input != nil {
		for k, v := range request.input {
			switch v.(type) {
			case string:
				val = attribute.StringValue(v.(string))
			case int:
				val = attribute.IntValue(v.(int))
			case int64:
				val = attribute.Int64Value(v.(int64))
			case float64:
				val = attribute.Float64Value(v.(float64))
			case bool:
				val = attribute.BoolValue(v.(bool))
			default:
				val = attribute.StringValue(fmt.Sprintf("%#v", v))
			}
			if val.Type() > 0 {
				attributes = append(attributes, attribute.KeyValue{
					Key:   attribute.Key("gen_ai.other_input." + k),
					Value: val,
				})
			}
			val = attribute.Value{}
		}
	}

	return attributes, parentContext
}

func (l LExperimentalAttributeExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request mcpRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	attributes, context = l.Base.OnEnd(attributes, context, request, response, err)
	if request.output != nil {
		var val attribute.Value
		for k, v := range request.output {
			switch v.(type) {
			case string:
				val = attribute.StringValue(v.(string))
			case int:
				val = attribute.IntValue(v.(int))
			case int64:
				val = attribute.Int64Value(v.(int64))
			case float64:
				val = attribute.Float64Value(v.(float64))
			case bool:
				val = attribute.BoolValue(v.(bool))
			default:
				val = attribute.StringValue(fmt.Sprintf("%#v", v))
			}
			if val.Type() > 0 {
				attributes = append(attributes, attribute.KeyValue{
					Key:   attribute.Key("gen_ai.other_output." + k),
					Value: val,
				})
			}
			val = attribute.Value{}
		}

	}
	return attributes, context
}

func BuildServerCommonOtelInstrumenter() instrumenter.Instrumenter[mcpRequest, any] {
	builder := instrumenter.Builder[mcpRequest, any]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[mcpRequest, any]{Getter: aiCommonRequest{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[mcpRequest]{}).
		AddAttributesExtractor(&LExperimentalAttributeExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MCP_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
func BuildClientCommonOtelInstrumenter() instrumenter.Instrumenter[mcpRequest, any] {
	builder := instrumenter.Builder[mcpRequest, any]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[mcpRequest, any]{Getter: aiCommonRequest{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[mcpRequest]{}).
		AddAttributesExtractor(&LExperimentalAttributeExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MCP_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
