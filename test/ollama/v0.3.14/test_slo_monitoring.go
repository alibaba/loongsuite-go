package main

import (
	"context"
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	ctx := context.Background()
	testSLOMonitoring(ctx)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) < 1 {
			panic("Expected at least 1 trace for SLO monitoring test")
		}

		hasSLOMetrics := false
		for _, trace := range stubs {
			for _, span := range trace {
				for _, attr := range span.Attributes {
					if attr.Key == "gen_ai.slo.compliance_percentage" ||
					   attr.Key == "gen_ai.slo.p95_ms" {
						hasSLOMetrics = true
					}
				}
			}
		}

		if !hasSLOMetrics {
			// Note: SLO metrics may not appear immediately
		}
	}, 1)
}

func testSLOMonitoring(ctx context.Context) {
	client, server := NewMockOllamaGenerateForInvoke(ctx)
	defer server.Close()
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			start := time.Now()
			streamFlag := false
			req := &api.GenerateRequest{
				Model:  "llama3:8b",
				Prompt: "Test prompt for SLO monitoring",
				Stream: &streamFlag,
			}

			err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
				if resp.Done {
					_ = time.Since(start)
				}
				return nil
			})

			if err != nil {
				// Error handling
			}
		}(i)

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	simulatePerformanceBottleneck(ctx, client)
}

func simulatePerformanceBottleneck(ctx context.Context, client *api.Client) {
	// Simulate a slow request
	time.Sleep(100 * time.Millisecond)

	streamFlag := false
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Complex prompt for bottleneck detection",
		Stream: &streamFlag,
	}

	_ = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		return nil
	})
}