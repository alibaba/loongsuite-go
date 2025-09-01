## Integrating with manual instrumentation

Automatic instrumentation already meets our needs in most scenarios, but manual instrumentation allows developers to have greater control over their projects.

### Automatic instrumentation

Based on the `example/demo`, automatic instrumentation generates a trace where the HTTP service acts as the root span, with Redis and MySQL operations as child spans.

![](images/auto_instr_jaeger.png)

### Combining with manual instrumentation

Manual instrumentation enables us to capture specific telemetry data. In `example/demo/pkg/http.go` we can add a manual span to the `traceService()` function that wraps database operations.

```go
var tracer = otel.Tracer("otel-manual-instr")

func traceService(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "db init")
	defer span.End()
    
    ...
}
```

And the generated trace in Jaeger is as follows.

![](images/manual_instr_jaeger.png)

