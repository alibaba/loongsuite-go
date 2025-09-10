# kratos-demo

## Overview [中文版](./README_CN.md)

This is a demonstration project showcasing loongsuite-go-agent integration with the Kratos framework.

## Architecture

The demo consists of two services:

- **WeatherService**: HTTP service listening on port 8080
- **MessageService**: gRPC service listening on port 8081

### Service Flow

1. WeatherService receives external HTTP requests
2. WeatherService calls MessageService via gRPC
3. MessageService processes and returns the result
4. WeatherService returns the result to the client

### Observability

The project uses loongsuite-go-agent to export traces and metrics to:
- **OpenTelemetry Collector**: Data collection and processing
- **Jaeger**: Distributed tracing
- **Prometheus**: Metrics storage

## Quick Start

### Running the Services

```shell
sh start.sh
```

### Testing the API

After startup, you can test the service with:

```shell
curl http://localhost:8080/v1/weather/{city}/message
```

**Example**:
```shell
curl http://localhost:8080/v1/weather/HangZhou/message
```

### Viewing Observability Data

- **Jaeger Tracing UI**: http://localhost:16686,after initiating the request, refresh to see the link tracking data.
- **Prometheus Metrics UI**: http://localhost:9090,Query '{exported_job="kratos-demo"}' to see the monitoring indicators of kratos-demo.

## Port Configuration

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| WeatherService | 8080 | HTTP | External REST API |
| MessageService | 8081 | gRPC | Internal service calls |
| Jaeger | 16686 | HTTP | Tracing UI |
| Prometheus | 9090 | HTTP | Metrics UI |
