// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package instrumenter

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"

	"go.opentelemetry.io/otel/attribute"
	ottrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// getScopeKey returns the appropriate span key based on scope name and span kind
func getScopeKey(scopeName string, spanKind trace.SpanKind) attribute.Key {
	metadata := utils.GetInstrumentationMetadata(scopeName)
	if metadata == nil {
		return ""
	}

	if spanKind == trace.SpanKindClient {
		return metadata.ClientKey
	}
	return metadata.ServerKey
}

type SpanSuppressor interface {
	StoreInContext(context context.Context, spanKind trace.SpanKind, span trace.Span) context.Context
	ShouldSuppress(parentContext context.Context, spanKind trace.SpanKind) bool
}

type NoopSpanSuppressor struct {
}

func NewNoopSpanSuppressor() *NoopSpanSuppressor {
	return &NoopSpanSuppressor{}
}

func (n *NoopSpanSuppressor) StoreInContext(context context.Context, spanKind trace.SpanKind, span trace.Span) context.Context {
	return context
}

func (n *NoopSpanSuppressor) ShouldSuppress(parentContext context.Context, spanKind trace.SpanKind) bool {
	return false
}

type SpanKeySuppressor struct {
	spanKeys []attribute.Key
}

func NewSpanKeySuppressor(spanKeys []attribute.Key) *SpanKeySuppressor {
	return &SpanKeySuppressor{spanKeys: spanKeys}
}

func (s *SpanKeySuppressor) StoreInContext(ctx context.Context, spanKind trace.SpanKind, span trace.Span) context.Context {
	// do nothing
	return ctx
}

func (s *SpanKeySuppressor) ShouldSuppress(parentContext context.Context, spanKind trace.SpanKind) bool {
	for _, spanKey := range s.spanKeys {
		span := trace.SpanFromContext(parentContext)
		if s, ok := span.(ottrace.ReadOnlySpan); ok {
			instScopeName := s.InstrumentationScope().Name
			if instScopeName != "" {
				parentSpanKind := s.SpanKind()
				parentSpanKey := getScopeKey(instScopeName, parentSpanKind)
				if spanKey != parentSpanKey {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func NewSpanKindSuppressor() *SpanKindSuppressor {
	return &SpanKindSuppressor{}
}

func (s *SpanKindSuppressor) StoreInContext(context context.Context, spanKind trace.SpanKind, span trace.Span) context.Context {
	// do nothing
	return context
}

func (s *SpanKindSuppressor) ShouldSuppress(parentContext context.Context, spanKind trace.SpanKind) bool {
	span := trace.SpanFromContext(parentContext)
	if readOnlySpan, ok := span.(ottrace.ReadOnlySpan); ok {
		instScopeName := readOnlySpan.InstrumentationScope().Name
		if instScopeName != "" {
			// Now we compare the actual span kinds directly
			// since scope name no longer distinguishes client/server
			parentSpanKind := readOnlySpan.SpanKind()
			if spanKind != parentSpanKind {
				return false
			}
		}
	} else {
		return false
	}
	return true
}
