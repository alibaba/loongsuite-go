# OpenAI Monitoring Hooks

This package provides automatic monitoring and instrumentation for OpenAI API interactions in Go applications. It supports both the official OpenAI SDK (`github.com/openai/openai-go`) and the popular community SDK (`github.com/sashabaranov/go-openai`).

## Overview

The OpenAI monitoring hooks automatically capture telemetry data from OpenAI API calls, including:

- Request parameters (model, temperature, max tokens, etc.)
- Response metadata (tokens used, finish reasons, response IDs)
- Timing and performance metrics
- Error tracking
- Distributed tracing integration

## Features

- **Zero-code instrumentation**: No changes required to your application code
- **Multi-SDK support**: Works with both official and community OpenAI SDKs
- **GenAI compliance**: Follows OpenTelemetry GenAI semantic conventions
- **Rich metrics**: Captures detailed request/response attributes
- **Streaming support**: Monitors both regular and streaming completions
- **Extensible**: Easy to customize for additional monitoring needs

## Supported SDKs

### 1. Official OpenAI SDK
- **Package**: `github.com/openai/openai-go`
- **Versions**: 1.0.0 to 2.0.0
- **Functions monitored**:
  - `chat/completions.New()` - Regular chat completions
  - `chat/completions.NewStreaming()` - Streaming chat completions

### 2. Community OpenAI SDK
- **Package**: `github.com/sashabaranov/go-openai`
- **Versions**: 1.0.0 to 2.0.0
- **Functions monitored**:
  - `Client.CreateChatCompletion()` - Regular chat completions
  - `Client.CreateChatCompletionStream()` - Streaming chat completions

**Note**: This also works with forks like `github.com/meguminnnnnnnnn/go-openai`.

## Installation

The OpenAI monitoring hooks are automatically included when you build your application with the `otel` tool:

```bash
otel go build -o myapp main.go
```

## Configuration

### Enabling/Disabling Monitoring

By default, OpenAI monitoring is **enabled**. To disable it:

```bash
export OTEL_INSTRUMENTATION_OPENAI_ENABLED=false
```

To explicitly enable it (default behavior):

```bash
export OTEL_INSTRUMENTATION_OPENAI_ENABLED=true
```

### Environment Variables

The following OpenTelemetry environment variables affect the OpenAI monitoring:

- `OTEL_INSTRUMENTATION_OPENAI_ENABLED`: Enable/disable OpenAI monitoring (default: `true`)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint for trace and metric export
- `OTEL_SERVICE_NAME`: Service name for telemetry
- `OTEL_TRACES_SAMPLER`: Sampling strategy (e.g., `always_on`, `traceidratio`)

## Usage Examples

### Example 1: Official OpenAI SDK

```go
package main

import (
    "context"
    "fmt"
    "github.com/openai/openai-go"
    "github.com/openai/openai-go/option"
)

func main() {
    // No instrumentation code needed - it's automatic!
    client := openai.NewClient(
        option.WithAPIKey("your-api-key"),
    )
    
    ctx := context.Background()
    
    // This call will be automatically monitored
    completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
        Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("Hello, how are you?"),
        }),
        Model: openai.F("gpt-4"),
        Temperature: openai.Float(0.7),
        MaxTokens: openai.Int(100),
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(completion.Choices[0].Message.Content)
}
```

### Example 2: Community OpenAI SDK

```go
package main

import (
    "context"
    "fmt"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    // No instrumentation code needed - it's automatic!
    client := openai.NewClient("your-api-key")
    
    ctx := context.Background()
    
    // This call will be automatically monitored
    resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Hello, how are you?",
            },
        },
        Temperature: 0.7,
        MaxTokens:   100,
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(resp.Choices[0].Message.Content)
}
```

### Example 3: Streaming Completions

```go
package main

import (
    "context"
    "fmt"
    "io"
    openai "github.com/sashabaranov/go-openai"
)

func main() {
    client := openai.NewClient("your-api-key")
    ctx := context.Background()
    
    // Streaming calls are also automatically monitored
    stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Tell me a story",
            },
        },
        Stream: true,
    })
    
    if err != nil {
        panic(err)
    }
    defer stream.Close()
    
    for {
        response, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        
        fmt.Print(response.Choices[0].Delta.Content)
    }
}
```

## Captured Attributes

The following attributes are captured for each OpenAI API call:

### Request Attributes (GenAI Semantic Conventions)

- `gen_ai.operation.name`: Operation type (`chat` or `chat.stream`)
- `gen_ai.system`: Always `"openai"`
- `gen_ai.request.model`: Model name (e.g., `gpt-4`, `gpt-3.5-turbo`)
- `gen_ai.request.temperature`: Temperature parameter
- `gen_ai.request.max_tokens`: Maximum tokens to generate
- `gen_ai.request.top_p`: Top-p sampling parameter
- `gen_ai.request.frequency_penalty`: Frequency penalty
- `gen_ai.request.presence_penalty`: Presence penalty
- `gen_ai.request.stop_sequences`: Stop sequences
- `gen_ai.request.seed`: Random seed (if provided)
- `server.address`: API endpoint base URL

### Response Attributes

- `gen_ai.response.id`: Response ID from OpenAI
- `gen_ai.response.model`: Model used in response
- `gen_ai.response.finish_reasons`: Finish reasons for all choices
- `gen_ai.usage.input_tokens`: Number of input tokens
- `gen_ai.usage.output_tokens`: Number of output tokens
- `gen_ai.usage.total_tokens`: Total tokens used

### Experimental Attributes

- `gen_ai.request.message_count`: Number of messages in request
- `gen_ai.response.content_length`: Length of response content
- `error.type`: Error type if the call failed
- `error.message`: Error message if the call failed

## Metrics

The following metrics are collected:

- **Operation Latency**: Time taken for each OpenAI API call
- **Operation Duration**: Total duration including retries
- **Token Usage**: Input, output, and total tokens per request

## Architecture

The OpenAI monitoring hooks follow the architecture patterns established by the Eino and MCP frameworks:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Code                          │
│                  (Your OpenAI API calls)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              OpenAI SDK (official or community)              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼ (Hooks injected at compile time)
┌─────────────────────────────────────────────────────────────┐
│                 OpenAI Monitoring Hooks                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Request Extraction (reflection-based)               │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     ▼                                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  OpenAI Instrumenter (GenAI attributes)              │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     ▼                                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  OpenTelemetry Span & Metrics                        │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│           OpenTelemetry Collector / Backend                  │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Data Types** (`openai_data_type.go`):
   - `openaiRequest`: Captures request parameters
   - `openaiResponse`: Captures response data

2. **Instrumenter** (`openai_otel_instrumenter.go`):
   - Implements GenAI semantic conventions
   - Extracts and formats attributes
   - Creates spans and metrics

3. **Setup Files**:
   - `setup_official.go`: Hooks for official OpenAI SDK
   - `setup_community.go`: Hooks for community SDK
   - Uses Go's `//go:linkname` for compile-time injection

## Extending the Monitoring

### Adding Custom Attributes

You can extend the monitoring by modifying the `OpenAIExperimentalAttributeExtractor` in `openai_otel_instrumenter.go`:

```go
func (o OpenAIExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request openaiRequest) ([]attribute.KeyValue, context.Context) {
    attributes, parentContext = o.Base.OnStart(attributes, parentContext, request)
    
    // Add your custom attribute
    attributes = append(attributes, attribute.KeyValue{
        Key:   "custom.my_attribute",
        Value: attribute.StringValue("my_value"),
    })
    
    return attributes, parentContext
}
```

### Adding Support for New Functions

To monitor additional OpenAI SDK functions:

1. Add hook functions in `setup_official.go` or `setup_community.go`
2. Update `tool/data/rules/openai.json` with the new function mapping
3. Rebuild with `make`

## Troubleshooting

### No traces are being generated

1. Check that monitoring is enabled:
   ```bash
   echo $OTEL_INSTRUMENTATION_OPENAI_ENABLED
   ```

2. Verify your application was built with the `otel` tool:
   ```bash
   otel go build -o myapp main.go
   ```

3. Check that OpenTelemetry is properly configured:
   ```bash
   echo $OTEL_EXPORTER_OTLP_ENDPOINT
   ```

### Incomplete attribute data

Some attributes may be unavailable depending on the request/response:
- Not all parameters are required in OpenAI API calls
- The SDK may have different field structures across versions
- Check that you're using a supported SDK version

### Build errors

If you encounter build errors:
1. Ensure you're using Go 1.24.0 or later
2. Run `go mod tidy` in the OpenAI rules directory
3. Check that all dependencies are available

## Related Documentation

- [Eino Framework Integration](../eino/README.md)
- [MCP Framework Integration](../mcp/README.md)
- [OpenTelemetry GenAI Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
- [Official OpenAI SDK Documentation](https://github.com/openai/openai-go)
- [Community OpenAI SDK Documentation](https://github.com/sashabaranov/go-openai)

## Contributing

To contribute improvements to the OpenAI monitoring hooks:

1. Add tests in the `test/openai/` directory
2. Follow the existing code structure and patterns
3. Ensure compatibility with both supported SDKs
4. Update documentation for any new features

## License

Copyright (c) 2025 Alibaba Group Holding Ltd.

Licensed under the Apache License, Version 2.0.
