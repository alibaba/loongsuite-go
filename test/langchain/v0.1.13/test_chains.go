package main

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/prompts"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	llm := fake.NewFakeLLM([]string{"你好"})
	m := prompts.NewPromptTemplate("30的3次方是多少", []string{})
	ch := chains.NewLLMChain(llm, m)
	_, err := chains.Call(context.Background(), ch, map[string]any{})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		// Top-level chain should have Workflow span kind
		verifier.VerifyLLMCommonAttributesWithGenAISpanKind(stubs[0][0], "chains", "langchain", trace.SpanKindClient, "workflow")
	}, 1)
}
