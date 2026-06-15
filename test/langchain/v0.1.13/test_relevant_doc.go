package main

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/vectorstores"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	vec := vectorstores.ToRetriever(fakeVectorDb{}, 1)
	_, err := vec.GetRelevantDocuments(context.Background(), "123")
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMCommonAttributes(stubs[0][0], "relevantDocuments", "langchain", trace.SpanKindClient)
	}, 1)
}
