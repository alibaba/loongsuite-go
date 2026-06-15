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

package zerolog

import (
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/sdk/trace"
)

type zeroLogInnerEnabler struct {
	enabled bool
}

func (z zeroLogInnerEnabler) Enable() bool {
	return z.enabled
}

var zeroLogEnabler = zeroLogInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ZEROLOG_ENABLED") != "false"}

//go:linkname zeroLogWriteOnEnter github.com/rs/zerolog.zeroLogWriteOnEnter
func zeroLogWriteOnEnter(call api.CallContext, ce *zerolog.Event, msg string) {
	if !zeroLogEnabler.Enable() {
		return
	}
	traceId, spanId := trace.GetTraceAndSpanId()
	if traceId != "" && spanId != "" {
		cer := ce.Str("trace_id", traceId).Str("span_id", spanId)
		call.SetParam(0, cer)
	}
	return
}
