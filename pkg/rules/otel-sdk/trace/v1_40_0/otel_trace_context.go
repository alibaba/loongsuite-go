//go:build ignore

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

package trace_context

import (
	"fmt"

	trace "go.opentelemetry.io/otel/trace"
)

const maxSpans = 300

type traceContext struct {
	sw  *spanWrapper
	n   int
	lcs trace.Span
}

type spanWrapper struct {
	span trace.Span
	prev *spanWrapper
}

func (tc *traceContext) size() int {
	return tc.n
}

func (tc *traceContext) add(span trace.Span) bool {
	if tc.n > 0 {
		if tc.n >= maxSpans {
			return false
		}
	}
	wrapper := &spanWrapper{span, tc.sw}
	// local root span
	if tc.n == 0 {
		tc.lcs = span
	}
	tc.sw = wrapper
	tc.n++
	return true
}

//go:norace
func (tc *traceContext) tail() trace.Span {
	if tc.n == 0 {
		return nil
	} else {
		return tc.sw.span
	}
}

func (tc *traceContext) localRootSpan() trace.Span {
	if tc.n == 0 {
		return nil
	} else {
		return tc.lcs
	}
}

func (tc *traceContext) del(span trace.Span) {
	if tc.n == 0 {
		return
	}
	addr := &tc.sw
	cur := tc.sw
	for cur != nil {
		sc1 := cur.span.SpanContext()
		sc2 := span.SpanContext()
		if sc1.TraceID() == sc2.TraceID() && sc1.SpanID() == sc2.SpanID() {
			*addr = cur.prev
			tc.n--
			break
		}
		addr = &cur.prev
		cur = cur.prev
	}
}

func (tc *traceContext) clear() {
	tc.sw = nil
	tc.n = 0
	SetBaggageContainerToGLS(nil)
}

//go:norace
func (tc *traceContext) TakeSnapShot() interface{} {
	// take a deep copy to avoid reading & writing the same map at the same time
	if tc.n == 0 {
		return &traceContext{nil, 0, nil}
	}
	last := tc.tail()
	sw := &spanWrapper{last, nil}
	return &traceContext{sw, 1, nil}
}

func GetGLocalData(key string) interface{} {
	//todo set key into traceContext struct
	//t := getOrInitTraceContext()

	return nil
}

func SetGLocalData(key string, value interface{}) {
	t := getOrInitTraceContext()

	setTraceContext(t)
}

func getOrInitTraceContext() *traceContext {
	tc := GetTraceContextFromGLS()
	if tc == nil {
		newTc := &traceContext{nil, 0, nil}
		setTraceContext(newTc)
		return newTc
	} else {
		return tc.(*traceContext)
	}
}

func setTraceContext(tc *traceContext) {
	SetTraceContextToGLS(tc)
}

func traceContextAddSpan(span trace.Span) {
	tc := getOrInitTraceContext()
	if !tc.add(span) {
		fmt.Println("Failed to add span to TraceContext")
	}
}

func GetTraceAndSpanId() (string, string) {
	tc := GetTraceContextFromGLS()
	if tc == nil || tc.(*traceContext).tail() == nil {
		return "", ""
	}
	ctx := tc.(*traceContext).tail().SpanContext()
	return ctx.TraceID().String(), ctx.SpanID().String()
}

func traceContextDelSpan(span trace.Span) {
	ctx := getOrInitTraceContext()
	ctx.del(span)
}

func clearTraceContext() {
	getOrInitTraceContext().clear()
}

func SpanFromGLS() trace.Span {
	gls := GetTraceContextFromGLS()
	if gls == nil {
		return nil
	}
	return gls.(*traceContext).tail()
}

func LocalRootSpanFromGLS() trace.Span {
	gls := GetTraceContextFromGLS()
	if gls == nil {
		return nil
	}
	return gls.(*traceContext).lcs
}
