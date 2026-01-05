module github.com/alibaba/loongsuite-go-agent/pkg/rules/ollama

go 1.23.0

toolchain go1.24.1

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0
	github.com/ollama/ollama v0.3.14
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg
