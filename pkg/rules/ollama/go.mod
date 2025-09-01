module github.com/alibaba/loongsuite-go-agent/pkg/rules/ollama

go 1.23.0

toolchain go1.24.1

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0
	github.com/ollama/ollama v0.3.14
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
)

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg
