# OpenTelemetry SDK Configuration

In addition to automatic instrumentation, the `otel` tool injects configuration code to initialize the OpenTelemetry SDK when the application starts. The following environment variables can be used to change the behavior of the OpenTelemetry SDK.

- `OTEL_SERVICE_NAME`: Specifies the service name for your application.
- `OTEL_TRACES_EXPORTER`: Specifies the trace exporter. Supported values: `none`, `console`, `zipkin`, `otlp`. Multiple exporters can be specified using comma-separated values (e.g., `console,otlp`). The default is `otlp`.
- `OTEL_METRICS_EXPORTER`: Specifies the metrics exporter. Supported values: `none`, `console`, `prometheus`, `otlp`. Multiple exporters can be specified using comma-separated values (e.g., `console,otlp`). The default is `otlp`.
- `OTEL_EXPORTER_OTLP_PROTOCOL`: Specifies the OTLP protocol for both traces and metrics. Supported values: `http/protobuf` (default), `grpc`.
- `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`: Specifies the OTLP protocol for traces, overriding `OTEL_EXPORTER_OTLP_PROTOCOL`. Supported values: `http/protobuf` (default), `grpc`.
- `OTEL_EXPORTER_OTLP_ENDPOINT`: Specifies the common endpoint for OTLP exporters.
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: Specifies the endpoint for OTLP trace exporter.
- `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`: Specifies the endpoint for OTLP metrics exporter.
- `OTEL_EXPORTER_OTLP_HEADERS`: Specifies headers for all OTLP exporters (e.g., `key1=value1,key2=value2`).
- `OTEL_EXPORTER_PROMETHEUS_PORT`: Specifies the port for the Prometheus exporter when `OTEL_METRICS_EXPORTER` is set to `prometheus`. Defaults to `9464`.
- `OTEL_METRIC_EXPORT_INTERVAL`: Specifies the metric export interval for periodic metric readers in milliseconds. This is the standard OpenTelemetry SDK environment variable.
- `OTEL_METRIC_EXPORT_INTERVALS`: Specifies multiple metric export intervals for periodic metric readers in milliseconds, using comma-separated values (e.g., `1000,60000,3600000`). This is a LoongSuite extension. When set with valid values, it overrides `OTEL_METRIC_EXPORT_INTERVAL`. It does not apply to the Prometheus pull exporter. Each periodic export is marked with resource attribute `telemetry.metric.export.interval.ms` so collectors can process different periods separately. The attribute value is an integer, so collector routing conditions should use numeric comparison, for example `resource.attributes["telemetry.metric.export.interval.ms"] == 1000`. Multiple intervals increase application-side export work; with multiple exporters, the number of periodic readers is roughly `len(exporters) * len(intervals)`.
- `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`: Specifies the aggregation temporality preference for metrics (case-insensitive). Supported values:
  - `cumulative` (default): All instrument kinds use Cumulative temporality
  - `delta`: Counter, Asynchronous Counter, and Histogram use Delta temporality; UpDownCounter and Asynchronous UpDownCounter use Cumulative temporality
  - `lowmemory`: Synchronous Counter and Histogram use Delta temporality; other types use Cumulative temporality (low memory mode)
- `OTEL_TRACE_SAMPLER`: Specifies the trace sampler. A floating-point number between 0.0 and 1.0 sets a ratio-based sampler. Values <= 0 will never sample, and values >= 1 will always sample. The default is a parent-based sampler that always samples.
- `OTEL_INSTRUMENTATION_HTTP_EXCLUDE_PATHS`: Specifies a regular expression pattern to exclude URL paths from HTTP auto-instrumentation (e.g., `^/(ping|health|metrics)$`). Requests whose paths match the pattern will not generate spans. By default, no paths are excluded.
