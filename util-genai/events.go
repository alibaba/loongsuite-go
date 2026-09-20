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

package utilgenai

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
)

// emitLLMEvent emits a `gen_ai.client.inference.operation.details` log event for
// an LLM invocation. It is a no-op unless experimental mode and event emission
// are both enabled. Message content is included only when content capture in
// events is enabled (EVENT_ONLY or SPAN_AND_EVENT).
func emitLLMEvent(logger log.Logger, invocation *LLMInvocation) {
	if logger == nil || invocation == nil {
		return
	}
	if !IsExperimentalMode() || !ShouldEmitEvent() {
		return
	}

	var attrs []attribute.KeyValue
	attrs = append(attrs, GetLLMCommonAttributes(invocation)...)
	attrs = append(attrs, GetLLMRequestAttributes(invocation)...)
	attrs = append(attrs, GetLLMResponseAttributes(invocation)...)

	if ShouldCaptureContentInEvent() {
		if len(invocation.InputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIInputMessages, InputMessagesToJSON(invocation.InputMessages)))
		}
		if len(invocation.OutputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIOutputMessages, OutputMessagesToJSON(invocation.OutputMessages)))
		}
		if len(invocation.SystemInstruction) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAISystemInstructions, SystemInstructionToJSON(invocation.SystemInstruction)))
		}
		if len(invocation.ToolDefinitions) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIToolDefinitions, ToolDefinitionsToJSON(invocation.ToolDefinitions)))
		}
	}

	emitEvent(invocation.ctx, logger, EventGenAIInferenceOperationDetails, attrs)
}

// emitInvokeAgentEvent emits a `gen_ai.client.agent.invoke.operation.details`
// log event for an agent invocation, following the same gating rules as
// emitLLMEvent.
func emitInvokeAgentEvent(logger log.Logger, invocation *InvokeAgentInvocation) {
	if logger == nil || invocation == nil {
		return
	}
	if !IsExperimentalMode() || !ShouldEmitEvent() {
		return
	}

	attrs := []attribute.KeyValue{
		GenAIOperationName(OperationInvokeAgent),
		GenAISpanKind(SpanKindAgent),
	}
	if invocation.AgentName != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentName, invocation.AgentName))
	}
	if invocation.AgentID != "" {
		attrs = append(attrs, attribute.String(AttrGenAIAgentID, invocation.AgentID))
	}
	if invocation.Provider != "" {
		attrs = append(attrs, GenAIProviderName(invocation.Provider))
	}
	if invocation.ConversationID != "" {
		attrs = append(attrs, GenAIConversationID(invocation.ConversationID))
	}

	if ShouldCaptureContentInEvent() {
		if len(invocation.InputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIInputMessages, InputMessagesToJSON(invocation.InputMessages)))
		}
		if len(invocation.OutputMessages) > 0 {
			attrs = append(attrs, attribute.String(AttrGenAIOutputMessages, OutputMessagesToJSON(invocation.OutputMessages)))
		}
	}

	emitEvent(invocation.ctx, logger, EventGenAIAgentInvokeOperationDetails, attrs)
}

// emitEvent builds and emits a log record with the given event name and
// attributes. The invocation context is used so the event is correlated with
// the active span.
func emitEvent(ctx context.Context, logger log.Logger, eventName string, attrs []attribute.KeyValue) {
	if ctx == nil {
		ctx = context.Background()
	}

	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetEventName(eventName)
	record.AddAttributes(attrs...)

	logger.Emit(ctx, record)
}
