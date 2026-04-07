module github.com/alibaba/loongsuite-go-agent/pkg/rules/otel-sdk/trace/span

go 1.24.0

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel/trace v1.39.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
)
