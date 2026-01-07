module test-openai-community

go 1.24.0

replace github.com/alibaba/loongsuite-go-agent => ../../../

require (
	github.com/alibaba/loongsuite-go-agent v0.0.0
	github.com/sashabaranov/go-openai v1.36.1
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
)
