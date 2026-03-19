// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package utilgenai provides OpenTelemetry utilities for GenAI instrumentation.

This package includes boilerplate and helpers to standardize instrumentation
for Generative AI. It provides APIs and types to minimize the work needed to
instrument GenAI libraries, while providing standardization for generating
OpenTelemetry spans, metrics, and events.

# Environment Variables

This package relies on environment variables to configure capturing of message content.
By default, message content will not be captured.

Set the environment variable OTEL_SEMCONV_STABILITY_OPT_IN to "gen_ai_latest_experimental"
to enable experimental features.

Set the environment variable OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT to one of:
  - "NO_CONTENT": Do not capture message content (default)
  - "SPAN_ONLY": Capture message content in spans only
  - "EVENT_ONLY": Capture message content in events only
  - "SPAN_AND_EVENT": Capture message content in both spans and events

# Span Attributes

This package provides these span attributes:
  - gen_ai.provider.name: Provider name (e.g., "openai")
  - gen_ai.operation.name: Operation name (e.g., "chat")
  - gen_ai.request.model: Request model name
  - gen_ai.response.finish_reasons: List of finish reasons
  - gen_ai.response.model: Response model name
  - gen_ai.response.id: Response ID
  - gen_ai.usage.input_tokens: Input token count
  - gen_ai.usage.output_tokens: Output token count
  - gen_ai.input.messages: Input messages (when content capturing is enabled)
  - gen_ai.output.messages: Output messages (when content capturing is enabled)
  - gen_ai.system_instructions: System instructions (when provided)

# Usage Example

	handler := utilgenai.GetTelemetryHandler()

	// Create an invocation object with your request data
	invocation := &utilgenai.LLMInvocation{
		RequestModel:  "gpt-4",
		Provider:      "openai",
		InputMessages: []utilgenai.InputMessage{...},
	}

	// Use the handler to manage the lifecycle of an LLM invocation
	handler.StartLLM(invocation)
	defer func() {
		if err != nil {
			handler.FailLLM(invocation, &utilgenai.Error{Message: err.Error(), Type: "MyError"})
		} else {
			handler.StopLLM(invocation)
		}
	}()

	// Make the actual LLM call and populate outputs
	response, err := client.Chat(ctx, request)
	invocation.OutputMessages = []utilgenai.OutputMessage{...}
	invocation.InputTokens = response.Usage.InputTokens
	invocation.OutputTokens = response.Usage.OutputTokens
*/
package utilgenai
