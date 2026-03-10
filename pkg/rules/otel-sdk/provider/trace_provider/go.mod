module github.com/alibaba/loongsuite-go-agent/pkg/rules/otel-sdk/provider/trace_provider

go 1.24.0

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel/trace v1.39.0
)


