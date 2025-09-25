package main

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testEmbeddings(ctx)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) < 2 {
			panic("Expected at least 2 traces for embeddings test")
		}

		hasEmbedding := false
		for _, trace := range stubs {
			for _, span := range trace {
				for _, attr := range span.Attributes {
					if attr.Key == "gen_ai.operation.name" {
						if attr.Value.AsString() == "embed" || attr.Value.AsString() == "embeddings" {
							hasEmbedding = true
						}
					}
				}
			}
		}

		if !hasEmbedding {
			panic("Embedding operations not properly traced")
		}
	}, 2)
}

func testEmbeddings(ctx context.Context) {
	// Test Embed API
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	embedReq := &api.EmbedRequest{
		Model: "llama3:8b",
		Input: "This is a test embedding input for observability",
	}

	// Mock response with embeddings
	embedResp, err := client.Embed(ctx, embedReq)
	if err == nil && embedResp != nil {
		// Successfully mocked
	}

	// Test Embeddings API
	embeddingsReq := &api.EmbeddingRequest{
		Model:  "llama3:8b",
		Prompt: "Another test for batch embeddings",
	}

	embeddingsResp, err := client.Embeddings(ctx, embeddingsReq)
	if err == nil && embeddingsResp != nil {
		// Successfully mocked
	}
}

