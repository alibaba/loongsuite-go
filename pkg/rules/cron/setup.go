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

package cron

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	cron "github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname cronRunOnEnter github.com/robfig/cron/v3.cronRunOnEnter
func cronRunOnEnter(call api.CallContext, c *cron.Cron) {
	if c == nil {
		return
	}
	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindServer)}
	_, span := otel.Tracer("github.com/robfig/cron/v3").Start(context.Background(), "cron", opts...)
	temp := make(map[string]interface{}, 1)
	temp["span"] = span
	call.SetData(temp)
}

//go:linkname cronRunOnExit github.com/robfig/cron/v3.cronRunOnExit
func cronRunOnExit(call api.CallContext) {
	if call.GetData() == nil {
		return
	}
	temp, ok := call.GetData().(map[string]interface{})
	if !ok || temp == nil {
		return
	}
	span, _ := temp["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	span.End()
}

//go:linkname funcJobRunOnEnter github.com/robfig/cron/v3.funcJobRunOnEnter
func funcJobRunOnEnter(call api.CallContext, f cron.FuncJob) {
	if f == nil {
		return
	}
	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindServer)}
	_, span := otel.Tracer("github.com/robfig/cron/v3").Start(context.Background(), "func cron", opts...)
	temp := make(map[string]interface{}, 1)
	temp["span"] = span
	call.SetData(temp)
}

//go:linkname funcJobRunOnExit github.com/robfig/cron/v3.funcJobRunOnExit
func funcJobRunOnExit(call api.CallContext) {
	if call.GetData() == nil {
		return
	}
	temp, ok := call.GetData().(map[string]interface{})
	if !ok || temp == nil {
		return
	}
	span, _ := temp["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	span.End()
}
