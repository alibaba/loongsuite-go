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
	"encoding/base64"
	"encoding/json"
	"reflect"
)

// GenAIJSONEncoder is a custom JSON encoder for GenAI data that handles
// special types like byte slices (encoding them as base64).
type GenAIJSONEncoder struct{}

// genAIValue is a wrapper to implement custom JSON marshaling.
type genAIValue struct {
	value any
}

// MarshalJSON implements json.Marshaler for genAIValue.
func (v genAIValue) MarshalJSON() ([]byte, error) {
	return marshalValue(v.value)
}

func marshalValue(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// Handle []byte as base64
			bytes := rv.Bytes()
			encoded := base64.StdEncoding.EncodeToString(bytes)
			return json.Marshal(encoded)
		}
		// Handle other slices
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = genAIValue{rv.Index(i).Interface()}
		}
		return json.Marshal(result)
	case reflect.Map:
		result := make(map[string]any)
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			result[key] = genAIValue{iter.Value().Interface()}
		}
		return json.Marshal(result)
	case reflect.Struct:
		return json.Marshal(v)
	default:
		return json.Marshal(v)
	}
}

// GenAIJSONDump serializes the value to JSON with GenAI-specific handling.
// It handles bytes as base64 encoding and uses compact JSON format.
func GenAIJSONDump(v any) ([]byte, error) {
	return marshalValue(v)
}

// GenAIJSONDumps serializes the value to a JSON string with GenAI-specific handling.
func GenAIJSONDumps(v any) (string, error) {
	data, err := GenAIJSONDump(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustJSONDumps serializes the value to a JSON string, returning empty string on error.
func MustJSONDumps(v any) string {
	s, _ := GenAIJSONDumps(v)
	return s
}

// MessagePartToMap converts a MessagePart to a map for JSON serialization.
func MessagePartToMap(part MessagePart) map[string]any {
	result := make(map[string]any)
	result["type"] = part.PartType()

	switch p := part.(type) {
	case Text:
		result["content"] = p.Content
	case Reasoning:
		result["content"] = p.Content
	case ToolCall:
		result["name"] = p.Name
		result["arguments"] = p.Arguments
		if p.ID != "" {
			result["id"] = p.ID
		}
	case ToolCallResponse:
		result["response"] = p.Response
		if p.ID != "" {
			result["id"] = p.ID
		}
	case Blob:
		result["modality"] = p.Modality
		result["content"] = base64.StdEncoding.EncodeToString(p.Content)
		if p.MimeType != "" {
			result["mime_type"] = p.MimeType
		}
	case File:
		result["modality"] = p.Modality
		result["file_id"] = p.FileID
		if p.MimeType != "" {
			result["mime_type"] = p.MimeType
		}
	case Uri:
		result["modality"] = p.Modality
		result["uri"] = p.URI
		if p.MimeType != "" {
			result["mime_type"] = p.MimeType
		}
	}

	return result
}

// InputMessageToMap converts an InputMessage to a map for JSON serialization.
func InputMessageToMap(msg InputMessage) map[string]any {
	parts := make([]map[string]any, len(msg.Parts))
	for i, part := range msg.Parts {
		parts[i] = MessagePartToMap(part)
	}
	return map[string]any{
		"role":  msg.Role,
		"parts": parts,
	}
}

// OutputMessageToMap converts an OutputMessage to a map for JSON serialization.
func OutputMessageToMap(msg OutputMessage) map[string]any {
	parts := make([]map[string]any, len(msg.Parts))
	for i, part := range msg.Parts {
		parts[i] = MessagePartToMap(part)
	}
	result := map[string]any{
		"role":  msg.Role,
		"parts": parts,
	}
	if msg.FinishReason != "" {
		result["finish_reason"] = msg.FinishReason
	}
	return result
}

// InputMessagesToJSON converts a slice of InputMessages to a JSON string.
func InputMessagesToJSON(messages []InputMessage) string {
	if len(messages) == 0 {
		return ""
	}
	maps := make([]map[string]any, len(messages))
	for i, msg := range messages {
		maps[i] = InputMessageToMap(msg)
	}
	return MustJSONDumps(maps)
}

// OutputMessagesToJSON converts a slice of OutputMessages to a JSON string.
func OutputMessagesToJSON(messages []OutputMessage) string {
	if len(messages) == 0 {
		return ""
	}
	maps := make([]map[string]any, len(messages))
	for i, msg := range messages {
		maps[i] = OutputMessageToMap(msg)
	}
	return MustJSONDumps(maps)
}

// SystemInstructionToJSON converts a slice of MessageParts to a JSON string.
func SystemInstructionToJSON(parts []MessagePart) string {
	if len(parts) == 0 {
		return ""
	}
	maps := make([]map[string]any, len(parts))
	for i, part := range parts {
		maps[i] = MessagePartToMap(part)
	}
	return MustJSONDumps(maps)
}

// ToolDefinitionsToJSON converts a slice of FunctionToolDefinitions to a JSON string.
func ToolDefinitionsToJSON(tools []FunctionToolDefinition) string {
	if len(tools) == 0 {
		return ""
	}
	return MustJSONDumps(tools)
}
