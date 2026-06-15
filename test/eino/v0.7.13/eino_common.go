// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func NewMockReActAgentLambda(ctx context.Context) (lba *compose.Lambda, err error) {
	ins, err := NewMockReactAgent(ctx)
	if err != nil {
		return nil, err
	}
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}

func NewMockReactAgent(ctx context.Context) (*react.Agent, error) {
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBodyWithToolCall := `{"id":"e849dd51-7887-4e96-ac76-01e0b6dfa8b6","object":"chat.completion","created":1752200031,"model":"mock-chat","choices":[{"index":0,"message":{"role":"assistant","tool_calls":[{"index":0,"id":"call_0_b1bbef1e-376f-475c-b569-3f8d6da5fa44","type":"function","function":{"name":"greet","arguments":"{\"name\": \"test user\"}"}}]},"finish_reason":"tool_calls","content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false},"jailbreak":{"filtered":false,"detected":false},"profanity":{"filtered":false,"detected":false}}}],"usage":{"prompt_tokens":124,"completion_tokens":19,"total_tokens":143,"prompt_tokens_details":{"audio_tokens":0,"cached_tokens":64},"completion_tokens_details":null},"system_fingerprint":"fp_8802369eaa_prod0623_fp8_kvcache"}`
		mockBodyFinal := `{"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion","created":1752200038,"model":"mock-chat","choices":[{"index":0,"message":{"role":"assistant","content":"Hello form mock tool"},"finish_reason":"stop","content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false},"jailbreak":{"filtered":false,"detected":false},"profanity":{"filtered":false,"detected":false}}}],"usage":{"prompt_tokens":1779,"completion_tokens":340,"total_tokens":2119,"prompt_tokens_details":{"audio_tokens":0,"cached_tokens":1728},"completion_tokens_details":null},"system_fingerprint":"fp_8802369eaa_prod0623_fp8_kvcache"}`
		if count == 0 {
			w.Write([]byte(mockBodyWithToolCall))
			count++
		} else if count == 1 {
			w.Write([]byte(mockBodyFinal))
			count++
		}
	}))

	config := &openai.ChatModelConfig{
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}

	mockTool := &MockGreetTool{}

	return react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{mockTool},
		},
		MaxStep: 3,
	})
}

func NewMockOllamaChatModelForInvoke(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBody := `{"model":"mock-chat","created_at":"2025-07-12T16:44:51.135504+08:00","message":{"role":"AI","content":"Hello! How can I assist you today?"},"done":true}`
		w.Write([]byte((mockBody)))
	}))
	config := &ollama.ChatModelConfig{
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = ollama.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockOllamaChatModelForStream(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`{"model":"mock-chat","created_at":"2025-07-12T16:44:51.135504+08:00","message":{"role":"assistant","content":"Hello"},"done":false}`,
			`{"model":"mock-chat","created_at":"2025-07-12T16:44:51.135504+08:00","message":{"role":"assistant","content":"! How can"},"done":false}`,
			`{"model":"mock-chat","created_at":"2025-07-12T16:44:51.135504+08:00","message":{"role":"assistant","content":" I assist you today?"},"done":false}`,
			`{"model":"mock-chat","created_at":"2025-07-12T16:44:51.135504+08:00","message":{"role":"assistant","content":""},"done":true}`,
		}

		for range chunks {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))

	config := &ollama.ChatModelConfig{
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = ollama.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockOpenAIChatModelForInvoke(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBody := `{"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion","created":1752200038,"model":"mock-chat","choices":[{"index":0,"message":{"role":"assistant","content":"Hello! How can I assist you today?"},"finish_reason":"stop","content_filter_results":{"hate":{"filtered":false},"self_harm":{"filtered":false},"sexual":{"filtered":false},"violence":{"filtered":false},"jailbreak":{"filtered":false,"detected":false},"profanity":{"filtered":false,"detected":false}}}],"usage":{"prompt_tokens":1779,"completion_tokens":340,"total_tokens":2119,"prompt_tokens_details":{"audio_tokens":0,"cached_tokens":1728},"completion_tokens_details":null},"system_fingerprint":"fp_8802369eaa_prod0623_fp8_kvcache"}`
		w.Write([]byte((mockBody)))
	}))
	config := &openai.ChatModelConfig{
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockOpenAIChatModelForStream(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion.chunk","created":1752200038,"model":"mock-chat","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion.chunk","created":1752200038,"model":"mock-chat","choices":[{"index":0,"delta":{"content":"! How can"},"finish_reason":null}]}`,
			`data: {"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion.chunk","created":1752200038,"model":"mock-chat","choices":[{"index":0,"delta":{"content":" I assist you today?"},"finish_reason":null}]}`,
			`data: {"id":"5f511bcc-d863-47e5-97a0-732d2a962c0f","object":"chat.completion.chunk","created":1752200038,"model":"mock-chat","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for range chunks {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))

	config := &openai.ChatModelConfig{
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockArkChatModelForInvoke(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBody := `{
  "choices": [
    {
      "finish_reason": "stop",
      "index": 0,
      "logprobs": null,
      "message": {
        "content": "Hello! How can I help you today?",
        "role": "assistant"
      }
    }
  ],
  "created": 1742631811,
  "id": "0217426318107460cfa43dc3f3683b1de1c09624ff49085a456ac",
  "model": "mock-chat",
  "service_tier": "default",
  "object": "chat.completion",
  "usage": {
    "completion_tokens": 9,
    "prompt_tokens": 19,
    "total_tokens": 28,
    "prompt_tokens_details": {
      "cached_tokens": 0
    },
    "completion_tokens_details": {
      "reasoning_tokens": 0
    }
  }
}`
		w.Write([]byte((mockBody)))
	}))
	config := &ark.ChatModelConfig{
		APIKey:  "mock-api",
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockArkChatModelForStream(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"Hello"},"finish_reason":null,"index":0}],"created":1742631811,"id":"stream-1","model":"mock-chat","object":"chat.completion.chunk"}`,
			`data: {"choices":[{"delta":{"content":"! How can"},"finish_reason":null,"index":0}],"created":1742631811,"id":"stream-1","model":"mock-chat","object":"chat.completion.chunk"}`,
			`data: {"choices":[{"delta":{"content":" I help you today?"},"finish_reason":null,"index":0}],"created":1742631811,"id":"stream-1","model":"mock-chat","object":"chat.completion.chunk"}`,
			`data: {"choices":[{"delta":{},"finish_reason":"stop","index":0}],"created":1742631811,"id":"stream-1","model":"mock-chat","object":"chat.completion.chunk","usage":{"completion_tokens":9,"prompt_tokens":19,"total_tokens":28}}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))

	config := &ark.ChatModelConfig{
		APIKey:  "mock-api",
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockClaudeChatModelForInvoke(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockBody := `{  
  "id": "msg_01XFDUDYJgAACzvnptvVoYEL",  
  "type": "message",  
  "role": "assistant",  
  "content": [  
    {  
      "type": "text",  
      "text": "Hello! How can I assist you today?"  
    }  
  ],  
  "model": "mock-chat",  
  "stop_reason": "end_turn",  
  "stop_sequence": null,  
  "usage": {  
    "input_tokens": 10,  
    "output_tokens": 25  
  }  
}`
		w.Write([]byte(mockBody))
	}))

	config := &claude.Config{
		APIKey:  "mock-api-key",
		BaseURL: &ts.URL,
		Model:   "mock-chat",
	}
	cm, err = claude.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockClaudeChatModelForStream(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`event: message_start  
data: {"type": "message_start", "message": {"id": "msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY", "type": "message", "role": "assistant", "content": [], "model": "claude-3-opus-20240229", "stop_reason": null, "stop_sequence": null, "usage": {"input_tokens": 25, "output_tokens": 1}}}`,
			`event: content_block_start  
data: {"type": "content_block_start", "index": 0, "content_block": {"type": "text", "text": ""}}`,
			`event: content_block_delta  
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}`,
			`event: content_block_delta  
data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "! How can I assist you today?"}}`,
			`event: content_block_stop  
data: {"type": "content_block_stop", "index": 0}`,
			`event: message_delta  
data: {"type": "message_delta", "delta": {"stop_reason": "end_turn", "stop_sequence":null}, "usage": {"output_tokens": 15}}`,
			`event: message_stop  
data: {"type": "message_stop"}`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))

	config := &claude.Config{
		APIKey:  "mock-api-key",
		BaseURL: &ts.URL,
		Model:   "mock-chat",
	}
	cm, err = claude.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockQwenChatModelForInvoke(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mockBody := `{  
  "id": "chatcmpl-9f8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c",  
  "object": "chat.completion",  
  "created": 1742631811,  
  "model": "mock-chat",
  "choices": [  
    {  
      "index": 0,  
      "message": {  
        "role": "assistant",  
        "content": "Hello! How can I assist you today?"  
      },  
      "finish_reason": "stop"  
    }  
  ],  
  "usage": {  
    "prompt_tokens": 19,  
    "completion_tokens": 9,  
    "total_tokens": 28  
  }  
}`
		w.Write([]byte(mockBody))
	}))

	config := &qwen.ChatModelConfig{
		APIKey:  "mock-api-key",
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = qwen.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func NewMockQwenChatModelForStream(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"id":"chatcmpl-9f8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c","object":"chat.completion.chunk","created":1742631811,"model":"mock-chat","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-9f8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c","object":"chat.completion.chunk","created":1742631811,"model":"mock-chat","choices":[{"index":0,"delta":{"content":"! How can"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-9f8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c","object":"chat.completion.chunk","created":1742631811,"model":"mock-chat","choices":[{"index":0,"delta":{"content":" I assist you today?"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-9f8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c","object":"chat.completion.chunk","created":1742631811,"model":"mock-chat","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))

	config := &qwen.ChatModelConfig{
		APIKey:  "mock-api-key",
		BaseURL: ts.URL,
		Model:   "mock-chat",
	}
	cm, err = qwen.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

type MockGreetTool struct{}

func (t *MockGreetTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "greet",
		Desc: "greet with name",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"name": {
					Desc:     "user name who to greet",
					Required: true,
					Type:     schema.String,
				},
			}),
	}, nil
}

func (t *MockGreetTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	return "Hello from mock tool!", nil
}

type MockLoader struct{}

func (m *MockLoader) Load(ctx context.Context, src document.Source, opts ...document.LoaderOption) ([]*schema.Document, error) {
	return []*schema.Document{
		{ID: "doc1", Content: "This is the content of the first document"},
		{ID: "doc2", Content: "This is the content of the second document"},
		{ID: "doc3", Content: "This is the content of the third document"},
	}, nil
}

type MockEmbedder struct{}

func (m *MockEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		embedding := make([]float64, 3)
		embedding[0] = float64(len(text))
		embedding[1] = float64(strings.Count(text, "文档"))
		embedding[2] = float64(i + 1)
		embeddings[i] = embedding
	}

	return embeddings, nil
}

type MockIndexer struct {
	storage map[string]*schema.Document
}

func NewMockIndexer() *MockIndexer {
	return &MockIndexer{
		storage: make(map[string]*schema.Document),
	}
}

func (m *MockIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	ids := make([]string, len(docs))

	for i, doc := range docs {
		m.storage[doc.ID] = doc
		ids[i] = doc.ID
	}

	return ids, nil
}

type MockRetriever struct {
	storage map[string]*schema.Document
}

func NewMockRetriever(storage map[string]*schema.Document) *MockRetriever {
	return &MockRetriever{storage: storage}
}

func (m *MockRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	var results []*schema.Document

	for _, doc := range m.storage {
		if strings.Contains(doc.Content, query) || strings.Contains(query, doc.ID) {
			results = append(results, doc)
		}
	}

	return results, nil
}
