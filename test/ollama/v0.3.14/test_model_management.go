package main

import (
	"context"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testModelManagement(ctx)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) < 1 {
			panic("Expected at least 1 trace for model management test")
		}

		hasModelOp := false
		for _, trace := range stubs {
			for _, span := range trace {
				for _, attr := range span.Attributes {
					if attr.Key == "gen_ai.model.operation" {
						hasModelOp = true
					}
				}
			}
		}

		if !hasModelOp {
			panic("Model management operations not properly traced")
		}
	}, 1)
}

func testModelManagement(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()

	// Test List API
	listResp, err := client.List(ctx)
	if err == nil && listResp != nil {
		// Successfully mocked
	}

	// Test Show API
	showReq := &api.ShowRequest{
		Model: "llama3:8b",
	}

	showResp, err := client.Show(ctx, showReq)
	if err == nil && showResp != nil {
		// Successfully mocked
	}

	// Test Copy API
	copyReq := &api.CopyRequest{
		Source:      "llama3:8b",
		Destination: "llama3-copy",
	}

	err = client.Copy(ctx, copyReq)
	if err == nil {
		// Successfully mocked
	}

	// Test Delete API
	deleteReq := &api.DeleteRequest{
		Model: "llama3-copy",
	}

	err = client.Delete(ctx, deleteReq)
	if err == nil {
		// Successfully mocked
	}

	// Test Pull API
	pullReq := &api.PullRequest{
		Model: "llama3:8b",
	}

	err = client.Pull(ctx, pullReq, func(progress api.ProgressResponse) error {
		return nil
	})
	if err == nil {
		// Successfully mocked
	}
}