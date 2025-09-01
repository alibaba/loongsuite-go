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

package log

import (
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel/sdk/trace"
)

type kitlogInnerEnabler struct {
	enabled bool
}

func (g kitlogInnerEnabler) Enable() bool {
	return g.enabled
}

var kitlogEnabler = kitlogInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GOKITLOG_ENABLED") != "false"}

//go:linkname logfmtLoggerLogOnEnter github.com/go-kit/log.logfmtLoggerLogOnEnter
func logfmtLoggerLogOnEnter(call api.CallContext, _ interface{}, keyVals ...interface{}) {
	if !kitlogEnabler.Enable() {
		return
	}

	traceId, spanId := trace.GetTraceAndSpanId()
	if traceId == "" && spanId == "" {
		return
	}

	newKeyVals := make([]interface{}, 0, len(keyVals)+4)
	newKeyVals = append(newKeyVals, keyVals...)
	if traceId != "" {
		newKeyVals = append(newKeyVals, "trace_id", traceId)
	}
	if spanId != "" {
		newKeyVals = append(newKeyVals, "span_id", spanId)
	}

	call.SetParam(1, newKeyVals)
}

//go:linkname jsonLoggerLogOnEnter github.com/go-kit/log.jsonLoggerLogOnEnter
func jsonLoggerLogOnEnter(call api.CallContext, _ interface{}, keyVals ...interface{}) {
	if !kitlogEnabler.Enable() {
		return
	}

	traceId, spanId := trace.GetTraceAndSpanId()
	if traceId == "" && spanId == "" {
		return
	}

	newKeyVals := make([]interface{}, 0, len(keyVals))
	newKeyVals = append(newKeyVals, keyVals...)
	if traceId != "" {
		newKeyVals = append(newKeyVals, "trace_id", traceId)
	}
	if spanId != "" {
		newKeyVals = append(newKeyVals, "span_id", spanId)
	}

	call.SetParam(1, newKeyVals)
}
