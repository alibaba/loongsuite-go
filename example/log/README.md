# How to inject traceID in log

## Automatic Injection

If we use the log framework supported by `loongsuite-go-agent`, TraceId and SpanId are automatically injected into the log.

```go
package main

import (
	"go.uber.org/zap"
	"net/http"
)

func main() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		logger := zap.NewExample()
		logger.Debug("this is debug message")
		logger.Info("this is info message")
		logger.Warn("this is warn message")
		logger.Error("this is error message")
	})

	http.ListenAndServe(":9999", nil)
}

```

For example, if we build the following Go file with `loongsuite-go-agent`, run the binary
and `curl localhost:9999/log`, we will
see the following output:

```shell
{"level":"debug","msg":"this is debug message","trace_id":"d62a8fea286cc66de9c68ca17d4faa88","span_id":"7cb6d692769ffd32"}
{"level":"info","msg":"this is info message","trace_id":"d62a8fea286cc66de9c68ca17d4faa88","span_id":"7cb6d692769ffd32"}
{"level":"warn","msg":"this is warn message","trace_id":"d62a8fea286cc66de9c68ca17d4faa88","span_id":"7cb6d692769ffd32"}
{"level":"error","msg":"this is error message","trace_id":"d62a8fea286cc66de9c68ca17d4faa88","span_id":"7cb6d692769ffd32"}
{"level":"debug","msg":"this is debug message","trace_id":"e56a6f1e7ed7af48cce8f64d045ed158","span_id":"def0b8cf10fe8844"}
{"level":"info","msg":"this is info message","trace_id":"e56a6f1e7ed7af48cce8f64d045ed158","span_id":"def0b8cf10fe8844"}
{"level":"warn","msg":"this is warn message","trace_id":"e56a6f1e7ed7af48cce8f64d045ed158","span_id":"def0b8cf10fe8844"}
{"level":"error","msg":"this is error message","trace_id":"e56a6f1e7ed7af48cce8f64d045ed158","span_id":"def0b8cf10fe8844"}
```

The TraceId and SpanId are automatically injected into the log.

## Manual Injection

If the framework is not supported by `loongsuite-go-agent`. We can manually inject TraceId and SpanId into the log:
```go
package main

import (
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	http.HandleFunc("/logwithtrace", func(w http.ResponseWriter, r *http.Request) {
		logger := zap.NewExample()
		traceId, spanId := trace.GetTraceAndSpanId()
		logger.Info("this is info message with fields",
			zap.String("traceId", traceId),
			zap.String("spanId", spanId),
		)
	})
	http.ListenAndServe(":9999", nil)
}
```

For example, if we build the following Go file with `loongsuite-go-agent`, run the binary and `curl localhost:9999/logwithtrace`, we will
see the following output:

```shell
{"level":"info","msg":"this is info message with fields","traceId":"92d63797010a2040484222a74c5ce304","spanId":"5a2c84c807a6e12c"}
```

The above code is placed in the [example/log](../example/log) directory
