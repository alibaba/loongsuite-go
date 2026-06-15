// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package ai

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// GenAISpanKindAttrsExtractor extracts the gen_ai.span.kind attribute from requests.
type GenAISpanKindAttrsExtractor[REQUEST any, RESPONSE any, GETTER GenAISpanKindGetter[REQUEST]] struct {
	Getter GETTER
}

func (e *GenAISpanKindAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	spanKind := e.Getter.GetGenAISpanKind(request)
	if spanKind == "" {
		spanKind = GenAISpanKindUnknown
	}
	attributes = append(attributes, spanKind.Attribute())
	return attributes, parentContext
}

func (e *GenAISpanKindAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, context
}
