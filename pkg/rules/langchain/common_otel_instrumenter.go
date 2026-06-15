package langchain

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type aiCommonRequest struct {
}

func (aiCommonRequest) GetAIOperationName(request langChainRequest) string {
	return request.operationName
}
func (aiCommonRequest) GetAISystem(request langChainRequest) string {
	return request.system
}
func (aiCommonRequest) GetGenAISpanKind(request langChainRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindUnknown
	}
	return request.spanKind
}

type LExperimentalAttributeExtractor struct {
	Base ai.AICommonAttrsExtractor[langChainRequest, any, aiCommonRequest]
}

func (l LExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request langChainRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = l.Base.OnStart(attributes, parentContext, request)

	spanKind := request.spanKind
	if spanKind == "" {
		spanKind = ai.GenAISpanKindUnknown
	}
	attributes = append(attributes, spanKind.Attribute())

	if request.input != nil {
		var val attribute.Value
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

func (l LExperimentalAttributeExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request langChainRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
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

func BuildCommonLangchainOtelInstrumenter() instrumenter.Instrumenter[langChainRequest, any] {
	builder := instrumenter.Builder[langChainRequest, any]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[langChainRequest, any]{Getter: aiCommonRequest{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[langChainRequest]{}).
		AddAttributesExtractor(&LExperimentalAttributeExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.LANGCHAIN_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
