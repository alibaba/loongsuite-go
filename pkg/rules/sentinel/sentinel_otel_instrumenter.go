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

package sentinel

import (
	"context"
	"time"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type sentinelInstrumenter struct {
	tracer trace.Tracer
}

func NewSentinelInstrumenter() *sentinelInstrumenter {
	tracer := otel.GetTracerProvider().
		Tracer(utils.SENTINEL_SCOPE_NAME,
			trace.WithInstrumentationVersion(version.Tag),
		)
	return &sentinelInstrumenter{
		tracer: tracer,
	}
}

func (s *sentinelInstrumenter) StartAndEnd(ctx context.Context, spanName string, StartTime time.Time, EndTime time.Time, attrs []attribute.KeyValue, opts ...trace.SpanStartOption) {
	// start and end span
	_, span := s.tracer.Start(context.Background(),
		spanName,
		trace.WithAttributes(attrs...),
		trace.WithTimestamp(StartTime),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	span.End(trace.WithTimestamp(EndTime))
}
