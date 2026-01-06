module otel

go 1.24.0



replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../../loongsuite-go-agent/test/verifier

replace github.com/alibaba/loongsuite-go-agent => ../../../loongsuite-go-agent

require go.opentelemetry.io/otel/trace v1.39.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
)
