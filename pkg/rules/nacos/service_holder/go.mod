module github.com/alibaba/loongsuite-go-agent/pkg/rules/nacos/service_holder

go 1.23.0

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	github.com/nacos-group/nacos-sdk-go/v2 v2.0.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/metric v1.35.0
)

require (
	github.com/go-errors/errors v1.0.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
