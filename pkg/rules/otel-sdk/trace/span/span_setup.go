package span

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	oTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	_ "unsafe"
)

//go:linkname nonRecordingSpanEndOnEnter go.opentelemetry.io/otel/sdk/trace.nonRecordingSpanEndOnEnter
func nonRecordingSpanEndOnEnter(call api.CallContext, span interface{}, options interface{}) {
	if span != nil {
		oTrace.TraceContextDelSpan(span.(trace.Span))
	}
}

//go:linkname recordingSpanEndOnEnter go.opentelemetry.io/otel/sdk/trace.recordingSpanEndOnEnter
func recordingSpanEndOnEnter(call api.CallContext, span interface{}, options interface{}) {
	if span != nil {
		oTrace.TraceContextDelSpan(span.(trace.Span))
	}
}
