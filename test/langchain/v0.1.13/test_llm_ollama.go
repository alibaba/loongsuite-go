package main

import (
	"bytes"
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"net/http/httptest"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBody := `{"model":"","created_at":"2025-03-12T16:44:51.135504+08:00","message":{"role":"AI","content":"ok"},"done":true}`
		w.Write(bytes.NewBufferString(mockBody).Bytes())
	}))
	defer ts.Close()
	opts := []ollama.Option{
		ollama.WithModel("deepseek-r1:8b"),
		//ollama.WithHTTPClient(httputil.DebugHTTPClient),// Debugging
		ollama.WithServerURL(ts.URL),
	}
	llm, err := ollama.New(opts...)
	if err != nil {
		panic(err)
	}

	_, err = llm.GenerateContent(context.Background(), []llms.MessageContent{llms.MessageContent{
		Role: "human",
		Parts: []llms.ContentPart{
			llms.TextPart("你好"),
		},
	}})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "deepseek-r1", "deepseek-r1:8b")
	}, 1)
}
