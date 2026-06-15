// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package ai

import "go.opentelemetry.io/otel/attribute"

// GenAISpanKind represents the type of span in a GenAI/LLM application.
type GenAISpanKind string

const (
	// GenAISpanKindWorkflow represents a workflow, typically the top-level span
	// that encompasses an entire LLM application or chain execution.
	GenAISpanKindWorkflow GenAISpanKind = "workflow"

	// GenAISpanKindTask represents a task, typically a sub-operation within a workflow.
	// This is the default type for chain operations that don't match other specific types.
	GenAISpanKindTask GenAISpanKind = "task"

	// GenAISpanKindAgent represents an agent, an autonomous entity that can make decisions
	// and execute actions based on LLM outputs.
	GenAISpanKindAgent GenAISpanKind = "agent"

	// GenAISpanKindTool represents a tool invocation, typically an external function
	// or API called by an agent or task.
	GenAISpanKindTool GenAISpanKind = "tool"

	// GenAISpanKindGeneration represents an LLM generation/completion operation.
	// This is used for direct LLM calls that generate text responses.
	GenAISpanKindGeneration GenAISpanKind = "generation"

	// GenAISpanKindEmbedding represents an embedding operation, where text is
	// converted to vector representations.
	GenAISpanKindEmbedding GenAISpanKind = "embedding"

	// GenAISpanKindRetriever represents a retriever operation, typically used
	// for document retrieval from vector stores or other sources.
	GenAISpanKindRetriever GenAISpanKind = "retriever"

	// GenAISpanKindReranker represents a reranker operation, used for
	// reordering/reranking retrieved documents based on relevance.
	// Note: Reserved for future use. Currently no dedicated reranker instrumentation in langchaingo.
	GenAISpanKindReranker GenAISpanKind = "reranker"

	// GenAISpanKindUnknown represents an unknown or unclassified span type.
	// This is used as a fallback when the span type cannot be determined.
	GenAISpanKindUnknown GenAISpanKind = "unknown"
)

// GenAISpanKindKey is the attribute key for the GenAI span kind.
const GenAISpanKindKey = attribute.Key("gen_ai.span.kind")

// String returns the string representation of the GenAISpanKind.
func (k GenAISpanKind) String() string {
	return string(k)
}

// Attribute returns an OpenTelemetry attribute for this span kind.
func (k GenAISpanKind) Attribute() attribute.KeyValue {
	return attribute.KeyValue{
		Key:   GenAISpanKindKey,
		Value: attribute.StringValue(string(k)),
	}
}

// GenAISpanKindGetter is an interface for extracting GenAI span kind from a request.
type GenAISpanKindGetter[REQUEST any] interface {
	GetGenAISpanKind(request REQUEST) GenAISpanKind
}
