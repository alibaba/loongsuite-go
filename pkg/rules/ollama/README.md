# Ollama Instrumentation

OpenTelemetry instrumentation for [Ollama](https://github.com/ollama/ollama).

## Features
- Chat and Generate API instrumentation
- Streaming support with TTFT tracking
- Token-based cost calculation
- Multi-currency support
- Budget monitoring with thresholds

## Configuration
- `OLLAMA_COST_CONFIG` - Custom pricing configuration file
- `OLLAMA_ENABLE_COST_TRACKING` - Enable/disable cost tracking (default: true)
- `OLLAMA_DEFAULT_CURRENCY` - Default currency for cost calculation (default: USD)