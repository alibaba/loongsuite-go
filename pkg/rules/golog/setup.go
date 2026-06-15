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
	"log"
	"os"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel/sdk/trace"
)

type glogInnerEnabler struct {
	enabled bool
}

func (g glogInnerEnabler) Enable() bool {
	return g.enabled
}

var glogEnabler = glogInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GLOG_ENABLED") != "false"}

//go:linkname goLogWriteOnEnter log.goLogWriteOnEnter
func goLogWriteOnEnter(call api.CallContext, ce *log.Logger, pc uintptr, calldepth int, appendOutput func([]byte) []byte) {
	if !glogEnabler.Enable() {
		return
	}
	traceId, spanId := trace.GetTraceAndSpanId()
	newAppendOutput := func(bytes []byte) []byte {
		sb := strings.Builder{}
		if traceId != "" {
			sb.WriteString(" trace_id=")
			sb.WriteString(traceId)
		}
		if spanId != "" {
			sb.WriteString(" span_id=")
			sb.WriteString(spanId)
		}
		bytes = append(bytes, []byte(sb.String())...)
		bytes = appendOutput(bytes)
		sb.Reset()
		return bytes
	}
	call.SetParam(3, newAppendOutput)
	return
}
