module rpcx-go-test

go 1.23.0

replace github.com/alibaba/loongsuite-go-agent => ../../../

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../../test/verifier

require (
	github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0-00010101000000-000000000000
	github.com/rpcx-ecosystem/rpcx-examples3 v1.1.6
	github.com/smallnest/rpcx v1.9.1
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/sdk/metric v1.39.0
)

