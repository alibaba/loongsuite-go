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

package zap

import (
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapInnerEnabler struct {
	enabled bool
}

func (z zapInnerEnabler) Enable() bool {
	return z.enabled
}

var zapEnabler = zapInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ZAP_ENABLED") != "false"}

//go:linkname zapLogWriteOnEnter go.uber.org/zap/zapcore.zapLogWriteOnEnter
func zapLogWriteOnEnter(call api.CallContext, ce *zapcore.CheckedEntry, fields ...zap.Field) {
	if !zapEnabler.Enable() {
		return
	}
	var traceIdOk, spanIdOk bool
	if fields != nil {
		for _, v := range fields {
			if v.Key == "trace_id" {
				traceIdOk = true
			} else if v.Key == "span_id" {
				spanIdOk = true
			}
		}
	}
	if !traceIdOk {
		traceId, spanId := trace.GetTraceAndSpanId()
		if traceId != "" {
			fields = append(fields, zap.String("trace_id", traceId))
		}
		if spanId != "" && !spanIdOk {
			fields = append(fields, zap.String("span_id", spanId))
		}
		call.SetParam(1, fields)
	}

	return
}
