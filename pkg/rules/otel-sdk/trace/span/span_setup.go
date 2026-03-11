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

// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package span

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	oTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	_ "unsafe"
)

//go:linkname nonRecordingSpanEndOnEnter go.opentelemetry.io/otel/sdk/trace.nonRecordingSpanEndOnEnter
func nonRecordingSpanEndOnEnter(call api.CallContext, span interface{}, options interface{}) {
	if span != nil {
		oTrace.TraceContextDelSpan(span.(trace.Span))
	}
}

//go:linkname recordingSpanEndOnEnter go.opentelemetry.io/otel/sdk/trace.recordingSpanEndOnEnter
func recordingSpanEndOnEnter(call api.CallContext, span interface{}, options interface{}) {
	if span != nil {
		oTrace.TraceContextDelSpan(span.(trace.Span))
	}
}
