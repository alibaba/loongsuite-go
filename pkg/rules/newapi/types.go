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

package newapi

import (
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const traceInfoCtxKey = "otel_stream_trace_info"

// newAPIInnerEnabler controls whether NewAPI instrumentation is enabled
type newAPIInnerEnabler struct {
	enabled bool
}

func (e newAPIInnerEnabler) Enable() bool {
	return e.enabled
}

var newAPIEnabler = newAPIInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_NEWAPI_ENABLED") != "false"}

var newAPITracer = otel.Tracer("loongsuite.instrumentation.newapi")

// streamTraceInfo carries trace information across hook points
type streamTraceInfo struct {
	Span         trace.Span
	Messages     []string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Model        string
}

// message represents a conversation entry for JSON serialization
type message struct {
	Role   string `json:"role"`
	Parts  []part `json:"parts"`
	Reason string `json:"finish_reason"`
}

// part represents message content for JSON serialization
type part struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
}
