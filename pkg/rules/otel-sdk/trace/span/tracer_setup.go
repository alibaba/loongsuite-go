package span

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	oTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	_ "unsafe"
)

//go:linkname newRecordingSpanOnExit go.opentelemetry.io/otel/sdk/trace.newRecordingSpanOnExit
func newRecordingSpanOnExit(call api.CallContext, span interface{}) {
	if span != nil {
		oTrace.TraceContextAddSpan(span.(trace.Span))
	}
}

//go:linkname newNonRecordingSpanOnExit go.opentelemetry.io/otel/sdk/trace.newNonRecordingSpanOnExit
func newNonRecordingSpanOnExit(call api.CallContext, span interface{}) {
	if span != nil {
		oTrace.TraceContextAddSpan(span.(trace.Span))
	}
}
