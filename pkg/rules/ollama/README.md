# Ollama Instrumentation

This package provides OpenTelemetry instrumentation for [Ollama](https://github.com/ollama/ollama), a framework for running large language models locally.

## Features

### Core Instrumentation
- **Chat API**: Full instrumentation for chat completions
- **Generate API**: Full instrumentation for text generation
- **OpenTelemetry Spans**: Automatic span creation with appropriate attributes
- **Token Metrics**: Capture input and output token counts

### Streaming Support
- **Real-time Streaming**: Full support for Ollama's callback-based streaming
- **TTFT Measurement**: Time To First Token tracking for streaming responses
- **Chunk Tracking**: Monitor streaming progress with chunk counts
- **Token Rate**: Calculate tokens per second throughput
- **Span Events**: Record streaming milestones (first token, progress, completion)

### Cost Calculation and Budget Monitoring (NEW)
- **Token-based Cost Calculation**: Automatic cost calculation based on model and token usage
- **Multi-Currency Support**: Built-in support for USD, EUR, CNY, GBP, JPY with automatic conversion
- **Configurable Pricing**: Customizable model pricing through environment variables or config files
- **Real-time Streaming Costs**: Incremental cost tracking during streaming responses
- **Budget Management**: Set and monitor spending limits with threshold alerts
- **SRE-style Error Budgets**: Track cost variance and detect anomalies
- **Sliding Window Monitoring**: Monitor costs over configurable time windows
- **Burn Rate Prediction**: Predict when budgets will be exhausted

## Instrumented Methods

The following Ollama API methods are instrumented:

- `Client.Generate()` - Text generation with optional streaming
- `Client.Chat()` - Chat completions with optional streaming

## Attributes Captured

### Standard OpenTelemetry GenAI Attributes
- `gen_ai.system`: "ollama"
- `gen_ai.request.model`: Model name (e.g., "llama3:8b")
- `gen_ai.operation.name`: "chat" or "generate"
- `gen_ai.usage.input_tokens`: Input token count
- `gen_ai.usage.output_tokens`: Output token count

### Streaming-Specific Attributes
- `gen_ai.response.streaming`: Boolean indicating if streaming was used
- `gen_ai.response.ttft_ms`: Time to first token in milliseconds
- `gen_ai.response.chunk_count`: Total number of streaming chunks
- `gen_ai.response.tokens_per_second`: Token generation throughput
- `gen_ai.response.stream_duration_ms`: Total streaming duration

### Cost Tracking Attributes
- `gen_ai.cost.input_tokens_usd`: Input token cost in USD
- `gen_ai.cost.output_tokens_usd`: Output token cost in USD
- `gen_ai.cost.total_usd`: Total cost in USD
- `gen_ai.cost.model_tier`: Pricing tier (economy/standard/premium)
- `gen_ai.cost.currency`: Currency used for calculation

### Budget Monitoring Attributes
- `gen_ai.budget.status`: Current budget status (ok/warning/critical/exceeded)
- `gen_ai.budget.usage_percentage`: Percentage of budget consumed
- `gen_ai.budget.threshold_exceeded`: Boolean if budget threshold exceeded
- `gen_ai.budget.anomaly_detected`: Boolean if cost anomaly detected
- `gen_ai.budget.error_budget_consumed`: Percentage of error budget used

## Streaming Behavior

Ollama streams by default when the `Stream` field is:
- `nil` (not specified) - **defaults to streaming**
- `true` - explicitly enables streaming
- `false` - explicitly disables streaming

## Span Events

For streaming requests, the following span events are recorded:

1. **First token received**: Records TTFT when first content chunk arrives
2. **Streaming progress**: Periodic updates every 10 chunks or 500ms
3. **Streaming completed**: Final metrics including total chunks and token rate

## Usage Examples

### Non-Streaming Request
```go
streamFlag := false
req := &api.GenerateRequest{
    Model:  "llama3:8b",
    Prompt: "Hello, world!",
    Stream: &streamFlag, // Explicitly disable streaming
}
```

### Streaming Request (Default)
```go
req := &api.GenerateRequest{
    Model:  "llama3:8b",
    Prompt: "Write a poem",
    // Stream not set - defaults to streaming
}
```

### Explicit Streaming Request
```go
streamFlag := true
req := &api.GenerateRequest{
    Model:  "llama3:8b",
    Prompt: "Write a poem",
    Stream: &streamFlag, // Explicitly enable streaming
}
```

## Cost Configuration

### Environment Variables

Cost tracking can be configured through environment variables:

```bash
# Enable/disable cost tracking (default: true)
export OLLAMA_ENABLE_COST_TRACKING=true

# Set default currency (default: USD)
export OLLAMA_DEFAULT_CURRENCY=USD

# Custom pricing configuration file
export OLLAMA_COST_CONFIG=/path/to/config.json
```

### Pricing Configuration File

Create a JSON file to customize model pricing:

```json
{
  "default_currency": "USD",
  "exchange_rates": {
    "USD": 1.0,
    "EUR": 0.85,
    "CNY": 7.25,
    "GBP": 0.79,
    "JPY": 149.50
  },
  "model_pricing": {
    "tinyllama": {
      "input_cost_per_1k": 0.00001,
      "output_cost_per_1k": 0.00002,
      "currency": "USD",
      "tier": "economy"
    },
    "llama3:8b": {
      "input_cost_per_1k": 0.00005,
      "output_cost_per_1k": 0.0001,
      "currency": "USD",
      "tier": "standard"
    },
    "llama3:70b": {
      "input_cost_per_1k": 0.0002,
      "output_cost_per_1k": 0.0004,
      "currency": "USD",
      "tier": "premium"
    }
  }
}
```

### Default Model Pricing

The following models have default pricing configured (per 1000 tokens):

| Model | Input Cost | Output Cost | Tier |
|-------|------------|-------------|------|
| tinyllama | $0.00001 | $0.00002 | economy |
| llama3:8b | $0.00005 | $0.0001 | standard |
| llama3:70b | $0.0002 | $0.0004 | premium |
| mistral:7b | $0.00004 | $0.00008 | standard |
| codellama:13b | $0.00008 | $0.00016 | standard |
| gemma:2b | $0.00002 | $0.00004 | economy |
| qwen:7b | $0.00003 | $0.00006 | standard |

## Testing

Run the instrumented tests with:

```bash
# Non-streaming tests
cd test/ollama/v0.3.14
../../../otel go build -o test_generate test_generate.go
../../../otel go build -o test_chat test_chat.go

# Streaming tests
../../../otel go build -o test_generate_stream test_generate_stream.go
../../../otel go build -o test_chat_stream test_chat_stream.go

# Cost calculation tests
../../../otel go build -o test_cost_calculation test_cost_calculation.go
../../../otel go build -o test_budget_tracking test_budget_tracking.go


# Backward compatibility test
../../../otel go build -o test_backward_compat test_backward_compat.go

# Run with OpenTelemetry export
OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318" ./test_generate_stream

# Run with cost tracking
OLLAMA_ENABLE_COST_TRACKING=true OLLAMA_DEFAULT_CURRENCY=USD ./test_cost_calculation
```

## Limitations

- **Input Token Counts**: Due to Ollama API design, input token counts are only available in the response, not the request
- **Streaming Token Counts**: Token counts are cumulative and only accurate in the final chunk
- **Model Parameters**: Advanced parameters (temperature, top_p, etc.) are not currently captured
- **Cost Accuracy**: Costs are estimated based on token counts and configurable pricing models
- **Budget Enforcement**: Budget limits are monitored but not enforced - applications must implement their own enforcement logic

## Version Support

- Minimum Ollama version: v0.3.14
- Maximum Ollama version: Latest (no upper limit)