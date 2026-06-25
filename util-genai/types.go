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

	"go.opentelemetry.io/otel/trace"
)

// ContentCapturingMode defines how message content should be captured.
type ContentCapturingMode int

const (
	// NoContent means do not capture content (default).
	NoContent ContentCapturingMode = iota
	// SpanOnly means only capture content in spans.
	SpanOnly
	// EventOnly means only capture content in events.
	EventOnly
	// SpanAndEvent means capture content in both spans and events.
	SpanAndEvent
)

// String returns the string representation of ContentCapturingMode.
func (m ContentCapturingMode) String() string {
	switch m {
	case NoContent:
		return "NO_CONTENT"
	case SpanOnly:
		return "SPAN_ONLY"
	case EventOnly:
		return "EVENT_ONLY"
	case SpanAndEvent:
		return "SPAN_AND_EVENT"
	default:
		return "NO_CONTENT"
	}
}

// ParseContentCapturingMode parses a string into ContentCapturingMode.
func ParseContentCapturingMode(s string) ContentCapturingMode {
	switch s {
	case "NO_CONTENT":
		return NoContent
	case "SPAN_ONLY":
		return SpanOnly
	case "EVENT_ONLY":
		return EventOnly
	case "SPAN_AND_EVENT":
		return SpanAndEvent
	default:
		return NoContent
	}
}

// OperationName defines the GenAI operation type.
type OperationName string

const (
	// Official GenAI operation names from the OpenTelemetry spec
	OperationChat            OperationName = "chat"
	OperationTextCompletion  OperationName = "text_completion"
	OperationGenerateContent OperationName = "generate_content"
	OperationEmbeddings      OperationName = "embeddings"
	OperationRetrieval       OperationName = "retrieval"
	OperationCreateAgent     OperationName = "create_agent"
	OperationInvokeAgent     OperationName = "invoke_agent"
	OperationExecuteTool     OperationName = "execute_tool"
	OperationInvokeWorkflow  OperationName = "invoke_workflow"
	OperationPlan            OperationName = "plan"
)

// FinishReason defines possible finish reasons for a generation.
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonError         FinishReason = "error"
)

// MessagePart represents a part of a message (text, tool call, etc.)
type MessagePart interface {
	PartType() string
}

// Text represents text content sent to or received from the model.
type Text struct {
	Content string `json:"content"`
}

func (t Text) PartType() string { return "text" }

// Reasoning represents reasoning/thinking content received from the model.
type Reasoning struct {
	Content string `json:"content"`
}

func (r Reasoning) PartType() string { return "reasoning" }

// ToolCall represents a tool call requested by the model.
type ToolCall struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Arguments any    `json:"arguments"`
}

func (t ToolCall) PartType() string { return "tool_call" }

// ToolCallResponse represents a tool call result sent to the model.
type ToolCallResponse struct {
	ID       string `json:"id,omitempty"`
	Response any    `json:"response"`
}

func (t ToolCallResponse) PartType() string { return "tool_call_response" }

// Modality represents the type of media content.
type Modality string

const (
	ModalityImage Modality = "image"
	ModalityVideo Modality = "video"
	ModalityAudio Modality = "audio"
)

// Blob represents blob binary data sent inline to the model.
type Blob struct {
	MimeType string   `json:"mime_type,omitempty"`
	Modality Modality `json:"modality"`
	Content  []byte   `json:"content"`
}

func (b Blob) PartType() string { return "blob" }

// File represents an external referenced file sent to the model by file id.
type File struct {
	MimeType string   `json:"mime_type,omitempty"`
	Modality Modality `json:"modality"`
	FileID   string   `json:"file_id"`
}

func (f File) PartType() string { return "file" }

// Uri represents an external referenced file sent to the model by URI.
type Uri struct {
	MimeType string   `json:"mime_type,omitempty"`
	Modality Modality `json:"modality"`
	URI      string   `json:"uri"`
}

func (u Uri) PartType() string { return "uri" }

// InputMessage represents an input message to the model.
type InputMessage struct {
	Role  string        `json:"role"`
	Parts []MessagePart `json:"parts"`
}

// OutputMessage represents an output message from the model.
type OutputMessage struct {
	Role         string        `json:"role"`
	Parts        []MessagePart `json:"parts"`
	FinishReason FinishReason  `json:"finish_reason,omitempty"`
}

// FunctionToolDefinition represents a function tool definition.
type FunctionToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

func (f FunctionToolDefinition) PartType() string { return "function" }

// LLMInvocation represents a single LLM call invocation.
// When creating an LLMInvocation object, only update the data attributes.
// The span and context attributes are set by the TelemetryHandler.
type LLMInvocation struct {
	// Request parameters
	RequestModel     string        `json:"request_model"`
	OperationName    OperationName `json:"operation_name"`
	Provider         string        `json:"provider,omitempty"`
	Temperature      *float64      `json:"temperature,omitempty"`
	TopP             *float64      `json:"top_p,omitempty"`
	TopK             *int          `json:"top_k,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	StopSequences    []string      `json:"stop_sequences,omitempty"`
	Seed             *int          `json:"seed,omitempty"`
	Stream           *bool         `json:"stream,omitempty"`
	ChoiceCount      *int          `json:"choice_count,omitempty"`
	OutputType       string        `json:"output_type,omitempty"`
	ReasoningLevel   string        `json:"reasoning_level,omitempty"`

	// Prompt template attributes
	PromptName    string `json:"prompt_name,omitempty"`
	PromptVersion string `json:"prompt_version,omitempty"`

	// Conversation tracking
	ConversationID        string `json:"conversation_id,omitempty"`
	ConversationCompacted *bool  `json:"conversation_compacted,omitempty"`

	// Messages
	InputMessages     []InputMessage           `json:"input_messages,omitempty"`
	OutputMessages    []OutputMessage          `json:"output_messages,omitempty"`
	SystemInstruction []MessagePart            `json:"system_instruction,omitempty"`
	ToolDefinitions   []FunctionToolDefinition `json:"tool_definitions,omitempty"`

	// Response data
	ResponseModelName string         `json:"response_model_name,omitempty"`
	ResponseID        string         `json:"response_id,omitempty"`
	FinishReasons     []FinishReason `json:"finish_reasons,omitempty"`
	InputTokens       *int           `json:"input_tokens,omitempty"`
	OutputTokens      *int           `json:"output_tokens,omitempty"`

	// Extended usage attributes (from official spec)
	ReasoningOutputTokens  *int `json:"reasoning_output_tokens,omitempty"`
	CacheReadInputTokens   *int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens *int `json:"cache_creation_input_tokens,omitempty"`

	// Streaming response metadata
	TimeToFirstChunk *float64 `json:"time_to_first_chunk,omitempty"`

	// Additional attributes
	Attributes map[string]any `json:"attributes,omitempty"`

	// Internal fields (managed by TelemetryHandler)
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
	monotonicEndS   float64
}

// NewLLMInvocation creates a new LLMInvocation with default values.
func NewLLMInvocation(requestModel string) *LLMInvocation {
	return &LLMInvocation{
		RequestModel:  requestModel,
		OperationName: OperationChat,
		Attributes:    make(map[string]any),
	}
}

// ============================================================================
// Extended Types (LoongSuite Extension)
//
// The following types are extensions beyond the basic LLM invocation support,
// providing instrumentation for embeddings, tool execution, agent operations,
// document retrieval, and reranking.
// ============================================================================

// EmbeddingInvocation represents an embedding operation invocation.
type EmbeddingInvocation struct {
	RequestModel    string   `json:"request_model"`
	Provider        string   `json:"provider,omitempty"`
	InputCount      *int     `json:"input_count,omitempty"`
	DimensionCount  *int     `json:"dimension_count,omitempty"`
	EncodingFormats []string `json:"encoding_formats,omitempty"`
	InputTokens     *int     `json:"input_tokens,omitempty"`

	// Response data
	ResponseModelName string `json:"response_model_name,omitempty"`

	Attributes map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewEmbeddingInvocation creates a new EmbeddingInvocation with default values.
func NewEmbeddingInvocation(requestModel string) *EmbeddingInvocation {
	return &EmbeddingInvocation{
		RequestModel: requestModel,
		Attributes:   make(map[string]any),
	}
}

// ExecuteToolInvocation represents a tool execution invocation.
type ExecuteToolInvocation struct {
	ToolName        string `json:"tool_name"`
	ToolCallID      string `json:"tool_call_id,omitempty"`
	ToolDescription string `json:"tool_description,omitempty"`
	ToolType        string `json:"tool_type,omitempty"`
	Input           any    `json:"input,omitempty"`
	Output          any    `json:"output,omitempty"`
	Attributes      map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewExecuteToolInvocation creates a new ExecuteToolInvocation.
func NewExecuteToolInvocation(toolName string) *ExecuteToolInvocation {
	return &ExecuteToolInvocation{
		ToolName:   toolName,
		Attributes: make(map[string]any),
	}
}

// InvokeAgentInvocation represents an agent invocation.
type InvokeAgentInvocation struct {
	AgentName        string         `json:"agent_name,omitempty"`
	AgentID          string         `json:"agent_id,omitempty"`
	AgentDescription string         `json:"agent_description,omitempty"`
	AgentVersion     string         `json:"agent_version,omitempty"`
	Provider         string         `json:"provider,omitempty"`
	ConversationID   string         `json:"conversation_id,omitempty"`
	InputMessages    []InputMessage `json:"input_messages,omitempty"`
	OutputMessages   []OutputMessage `json:"output_messages,omitempty"`
	Attributes       map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewInvokeAgentInvocation creates a new InvokeAgentInvocation.
func NewInvokeAgentInvocation() *InvokeAgentInvocation {
	return &InvokeAgentInvocation{
		Attributes: make(map[string]any),
	}
}

// CreateAgentInvocation represents an agent creation invocation.
type CreateAgentInvocation struct {
	AgentName        string         `json:"agent_name,omitempty"`
	AgentID          string         `json:"agent_id,omitempty"`
	AgentDescription string         `json:"agent_description,omitempty"`
	AgentVersion     string         `json:"agent_version,omitempty"`
	Provider         string         `json:"provider,omitempty"`
	Description      string         `json:"description,omitempty"`
	Attributes       map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewCreateAgentInvocation creates a new CreateAgentInvocation.
func NewCreateAgentInvocation() *CreateAgentInvocation {
	return &CreateAgentInvocation{
		Attributes: make(map[string]any),
	}
}

// RetrieveInvocation represents a document retrieval operation.
type RetrieveInvocation struct {
	QueryText      string         `json:"query_text,omitempty"`
	TopK           *int           `json:"top_k,omitempty"`
	DataSourceID   string         `json:"data_source_id,omitempty"`
	Provider       string         `json:"provider,omitempty"`
	Attributes     map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewRetrieveInvocation creates a new RetrieveInvocation.
func NewRetrieveInvocation() *RetrieveInvocation {
	return &RetrieveInvocation{
		Attributes: make(map[string]any),
	}
}

// RerankInvocation represents a document reranking operation.
// (LoongSuite Extension - reranking is not part of the official spec)
type RerankInvocation struct {
	Query         string         `json:"query,omitempty"`
	Model         string         `json:"model,omitempty"`
	Provider      string         `json:"provider,omitempty"`
	TopN          *int           `json:"top_n,omitempty"`
	InputCount    *int           `json:"input_count,omitempty"`
	OutputCount   *int           `json:"output_count,omitempty"`
	Attributes    map[string]any `json:"attributes,omitempty"`

	// Internal fields
	span            trace.Span
	ctx             context.Context
	cancelFunc      context.CancelFunc
	monotonicStartS float64
}

// NewRerankInvocation creates a new RerankInvocation.
func NewRerankInvocation() *RerankInvocation {
	return &RerankInvocation{
		Attributes: make(map[string]any),
	}
}

// Error represents an error that occurred during an invocation.
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
