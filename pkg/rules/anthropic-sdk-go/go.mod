module github.com/alibaba/loongsuite-go-agent/pkg/rules/anthropic-sdk-go

go 1.24

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	github.com/anthropics/anthropic-sdk-go v1.25.0
	go.opentelemetry.io/otel v1.40.0
)

require (
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg
