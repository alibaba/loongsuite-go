package main

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/embeddings"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	emb, err := embeddings.NewEmbedder(embedFake{},
		embeddings.WithBatchSize(15), // Batch upload limit
	)
	if err != nil {
		panic(err)
	}
	txt1 := `君不见黄河之水天上来，奔流到海不复回。`
	txt2 := `君不见高堂明镜悲白发，朝如青丝暮成雪。`
	_, err = emb.EmbedQuery(context.Background(), txt1)
	if err != nil {
		panic(err)
	}
	_, err = emb.EmbedDocuments(context.Background(), []string{txt1, txt2})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMCommonAttributes(stubs[0][0], "singleEmbed", "langchain", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[1][0], "batchedEmbed", "langchain", trace.SpanKindClient)
	}, 2)
}
