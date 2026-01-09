module test

go 1.22.0

require (
	github.com/alibaba/loongsuite-go-agent v0.0.0
	github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0
	github.com/openai/openai-go v3.0.0
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
)

replace github.com/alibaba/loongsuite-go-agent => ../../../

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../verifier
