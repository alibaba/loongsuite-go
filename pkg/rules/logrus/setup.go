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

package logrus

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
	"os"
)

type logrusInnerEnabler struct {
	enabled bool
}

func (l logrusInnerEnabler) Enable() bool {
	return l.enabled
}

var logrusEnabler = logrusInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_LOGRUS_ENABLED") != "false"}

//go:linkname withFieldOnExit github.com/sirupsen/logrus.withFieldOnExit
func withFieldOnExit(call api.CallContext, e *logrus.Entry) {
	if !logrusEnabler.Enable() {
		return
	}
	if e.Logger.Hooks == nil {
		e.Logger.Hooks = make(logrus.LevelHooks)
	}
	e.Logger.AddHook(&logHook{})
	return
}

//go:linkname logStandNewOnExit github.com/sirupsen/logrus.logStandNewOnExit
func logStandNewOnExit(call api.CallContext) {
	if !logrusEnabler.Enable() {
		return
	}
	std := logrus.StandardLogger()
	std.AddHook(&logHook{})
	return
}

//go:linkname logNewOnExit github.com/sirupsen/logrus.logNewOnExit
func logNewOnExit(call api.CallContext, log *logrus.Logger) {
	if !logrusEnabler.Enable() {
		return
	}
	log.AddHook(&logHook{})
	return
}

//go:linkname standardLoggerOnExit github.com/sirupsen/logrus.standardLoggerOnExit
func standardLoggerOnExit(call api.CallContext, log *logrus.Logger) {
	if !logrusEnabler.Enable() {
		return
	}
	log.AddHook(&logHook{})
	return
}

//go:linkname logNewOnEnter github.com/sirupsen/logrus.logNewOnEnter
func logNewOnEnter(call api.CallContext, log *logrus.Logger, formatter logrus.Formatter) {
	if !logrusEnabler.Enable() {
		return
	}
	call.SetData(log)
}

type logHook struct{}

func (hook *logHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *logHook) Fire(entry *logrus.Entry) error {
	if !logrusEnabler.Enable() {
		return nil
	}
	// Modify log content
	traceId, spanId := trace.GetTraceAndSpanId()
	if traceId != "" {
		entry.Data["trace_id"] = traceId
	}
	if spanId != "" {
		entry.Data["span_id"] = spanId
	}
	return nil
}
