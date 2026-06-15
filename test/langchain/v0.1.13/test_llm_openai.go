package main

import (
	"bytes"
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"net/http/httptest"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockBody := `{"choices":[{"index":0,"delta":{"role":"assistant","content":"hello"},"finish_reason":"stop"}]}`
		w.Write(bytes.NewBufferString(mockBody).Bytes())
	}))
	defer ts.Close()
	llm, err := openai.New(
		openai.WithModel("deepseek-reasoner"),
		openai.WithToken("token"),
		openai.WithBaseURL(ts.URL),
		openai.WithHTTPClient(httputil.DebugHTTPClient),
	)
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
		verifier.VerifyLLMAttributes(stubs[0][0], "chat", "deepseek-reasoner", "deepseek-reasoner")
	}, 1)
}
