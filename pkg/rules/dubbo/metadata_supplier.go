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

package dubbo

import (
	"context"
)

import (
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type dubboMetadataSupplier struct {
	metadata map[string]any
}

var _ propagation.TextMapCarrier = &dubboMetadataSupplier{}

func (s *dubboMetadataSupplier) Get(key string) string {
	if s.metadata == nil {
		return ""
	}
	item, ok := s.metadata[key].([]string)
	if !ok {
		return ""
	}
	if len(item) == 0 {
		return ""
	}
	return item[0]
}

func (s *dubboMetadataSupplier) Set(key string, value string) {
	if s.metadata == nil {
		s.metadata = map[string]any{}
	}
	s.metadata[key] = value
}

func (s *dubboMetadataSupplier) Keys() []string {
	out := make([]string, 0, len(s.metadata))
	for key := range s.metadata {
		out = append(out, key)
	}
	return out
}

func inject(ctx context.Context, metadata map[string]any, propagators propagation.TextMapPropagator) {
	propagators.Inject(ctx, &dubboMetadataSupplier{
		metadata: metadata,
	})
}

func extract(ctx context.Context, metadata map[string]any, propagators propagation.TextMapPropagator) (baggage.Baggage, trace.SpanContext) {
	ctx = propagators.Extract(ctx, &dubboMetadataSupplier{
		metadata: metadata,
	})
	return baggage.FromContext(ctx), trace.SpanContextFromContext(ctx)
}
