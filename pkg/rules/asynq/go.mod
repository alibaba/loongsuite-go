module github.com/alibaba/loongsuite-go-agent/pkg/rules/asynq

go 1.24

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	github.com/hibiken/asynq v0.23.0
	go.opentelemetry.io/otel v1.40.0
)
