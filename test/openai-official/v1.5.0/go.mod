module test-openai-official

go 1.24.0

replace github.com/alibaba/loongsuite-go-agent => ../../../

require (
github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0-20260107074919-08c36b668c42
github.com/openai/openai-go v1.5.0
go.opentelemetry.io/otel/sdk v1.39.0
)
