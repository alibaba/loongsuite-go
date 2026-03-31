module asynq/v0.26.0

go 1.24

replace github.com/alibaba/loongsuite-go-agent => ../../../

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../../test/verifier

require (
	github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0
	github.com/hibiken/asynq v0.26.0
	go.opentelemetry.io/otel/sdk v1.40.0
)
