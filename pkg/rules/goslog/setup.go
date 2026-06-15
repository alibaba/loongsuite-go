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

package golog

import (
	"context"
	"log/slog"
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel/sdk/trace"
)

type goSlogInnerEnabler struct {
	enabled bool
}

func (g goSlogInnerEnabler) Enable() bool {
	return g.enabled
}

var goSlogEnabler = goSlogInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GOSLOG_ENABLED") != "false"}

//go:linkname goSlogWriteOnEnter log/slog.goSlogWriteOnEnter
func goSlogWriteOnEnter(call api.CallContext, ce *slog.Logger, ctx context.Context, level slog.Level, msg string, args ...any) {
	if !goSlogEnabler.Enable() {
		return
	}
	traceId, spanId := trace.GetTraceAndSpanId()
	if traceId != "" {
		args = append(args, "trace_id", traceId)
	}
	if spanId != "" {
		args = append(args, "span_id", spanId)
	}
	call.SetParam(4, args)
	return
}
