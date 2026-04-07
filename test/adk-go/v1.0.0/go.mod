module test-adk-go

go 1.24.4

replace github.com/alibaba/loongsuite-go-agent => ../../../

require (
	github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0-20260107074919-08c36b668c42
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	google.golang.org/adk v1.0.0
	google.golang.org/genai v1.40.0
)
