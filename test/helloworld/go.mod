module helloworld

go 1.24.0

replace github.com/alibaba/loongsuite-go-agent => ../../

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../pkg

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../test/verifier

require (
	go.opentelemetry.io/otel v1.39.0
	golang.org/x/text v0.25.0
	golang.org/x/time v0.11.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
)
