# gRPC Demo

This demo shows how to use loongsuite-go to automatically instrument gRPC applications with OpenTelemetry. It includes both unary and server-streaming RPC examples.

## How to run it?

### 1. Build agent

Go to the root directory of `loongsuite-go` and execute:

```shell
make clean && make build
```

### 2. Do hybrid compilation

```shell
cd example/grpc-demo
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
OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4318" \
OTEL_EXPORTER_OTLP_INSECURE=true \
OTEL_SERVICE_NAME=grpc-demo \
./grpc-demo
```

### 5. Check trace data

Access Jaeger UI: http://localhost:16686

You should see traces for both the gRPC client and server spans, including:
- Unary RPC: `/HelloGrpc/Hello`
- Server Streaming RPC: `/HelloGrpc/StreamHello`
