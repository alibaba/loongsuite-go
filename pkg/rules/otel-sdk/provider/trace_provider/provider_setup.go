package span

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	_ "unsafe"
)

//go:linkname setTracerProviderOnEnter go.opentelemetry.io/otel.setTracerProviderOnEnter
func setTracerProviderOnEnter(call api.CallContext, tp trace.TracerProvider) {
	if otel.SetGlobalProviderEnable {
		call.SetSkipCall(true)
		return
	}
}
