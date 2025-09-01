module github.com/alibaba/loongsuite-go-agent/pkg/rules/gocql

go 1.23.0

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	github.com/gocql/gocql v1.3.0
	go.opentelemetry.io/otel/sdk v1.35.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)
