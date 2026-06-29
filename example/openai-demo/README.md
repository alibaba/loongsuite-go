# OpenAI Demo

This demo shows how to use loongsuite-go to automatically instrument OpenAI API calls with OpenTelemetry. It includes both standard and streaming chat completion examples.

## How to run it?

### 1. Build agent

Go to the root directory of `loongsuite-go` and execute:

```shell
make clean && make build
```

### 2. Do hybrid compilation

```shell
cd example/openai-demo
../../otel go build
```

### 3. Run Jaeger (optional, for viewing traces)

```shell
docker run --rm --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:1.53.0
```

### 4. Run the demo

```shell
OPENAI_API_KEY="your-api-key" \
OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4318" \
OTEL_EXPORTER_OTLP_INSECURE=true \
OTEL_SERVICE_NAME=openai-demo \
./openai-demo
```

You can also set `OPENAI_BASE_URL` to use a compatible API provider:

```shell
OPENAI_API_KEY="your-api-key" \
OPENAI_BASE_URL="https://your-api-provider/v1" \
OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4318" \
OTEL_EXPORTER_OTLP_INSECURE=true \
OTEL_SERVICE_NAME=openai-demo \
./openai-demo
```

### 5. Check trace data

Access Jaeger UI: http://localhost:16686

You should see traces for OpenAI API calls, including:
- Chat completion spans with model, token usage, and request/response attributes
- Streaming chat completion spans

| Environment Variable | Description                          | Example                        |
|---------------------|--------------------------------------|--------------------------------|
| OPENAI_API_KEY      | OpenAI API key (required)            | sk-xxx                         |
| OPENAI_BASE_URL     | Custom API base URL (optional)       | https://api.example.com/v1     |
