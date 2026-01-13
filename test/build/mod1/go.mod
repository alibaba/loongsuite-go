module m1

go 1.24.0

toolchain go1.24.11

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
)

require (
	github.com/alibaba/loongsuite-go-agent/pkg/rules/otel-context v0.0.0-00010101000000-000000000000
	github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt1 v0.0.0-00010101000000-000000000000
	github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt4 v0.0.0-00010101000000-000000000000
	github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt5 v0.0.0-00010101000000-000000000000
	github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt6 v0.0.0-00010101000000-000000000000
	github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt7 v0.0.0-00010101000000-000000000000
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.63.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.61.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.39.0

replace go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.39.0

replace go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.39.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0

replace go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.39.0

replace go.opentelemetry.io/otel/exporters/prometheus => go.opentelemetry.io/otel/exporters/prometheus v0.61.0

replace google.golang.org/protobuf => google.golang.org/protobuf v1.35.2

replace go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.39.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0

replace go.opentelemetry.io/otel/exporters/stdout/stdoutmetric => go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.39.0

replace go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.39.0

replace go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.39.0

replace go.opentelemetry.io/contrib/instrumentation/runtime => go.opentelemetry.io/contrib/instrumentation/runtime v0.63.0

replace go.opentelemetry.io/otel/exporters/zipkin => go.opentelemetry.io/otel/exporters/zipkin v1.39.0

replace go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.39.0

replace github.com/alibaba/loongsuite-go-agent/pkg => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/otel-context => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/otel-context

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt7 => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/test/fmt7

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt6 => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/test/fmt6

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt4 => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/test/fmt4

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt5 => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/test/fmt5

replace github.com/alibaba/loongsuite-go-agent/pkg/rules/test/fmt1 => /home/runner/work/loongsuite-go-agent/loongsuite-go-agent/test/build/.otel-build/alibaba-pkg/pkg/rules/test/fmt1
