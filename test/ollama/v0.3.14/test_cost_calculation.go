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
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {
	// Test cost calculation with tinyllama model
	fmt.Println("Testing Ollama cost calculation instrumentation...")
	fmt.Println("==================================================")

	// Set environment variables for cost tracking
	os.Setenv("OLLAMA_ENABLE_COST_TRACKING", "true")
	os.Setenv("OLLAMA_DEFAULT_CURRENCY", "USD")

	// Create client
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Test 1: Non-streaming request with cost calculation
	fmt.Println("\n1. Testing non-streaming request with cost calculation...")
	testNonStreamingWithCost(ctx, client)

	// Test 2: Streaming request with real-time cost accumulation
	fmt.Println("\n2. Testing streaming request with real-time cost accumulation...")
	testStreamingWithCost(ctx, client)

	// Test 3: Multiple requests to test budget tracking
	fmt.Println("\n3. Testing budget tracking with multiple requests...")
	testBudgetTracking(ctx, client)

	// Test 4: Currency conversion
	fmt.Println("\n4. Testing multi-currency support...")
	testCurrencyConversion(ctx, client)

	fmt.Println("\n==================================================")
	fmt.Println("Cost calculation tests completed!")
}

func testNonStreamingWithCost(ctx context.Context, client *api.Client) {
	req := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Count from 1 to 5",
		Stream: new(bool), // false - non-streaming
	}

	var finalResponse api.GenerateResponse
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Display cost information (would be in telemetry attributes)
	fmt.Printf("Response received (non-streaming)\n")
	fmt.Printf("Input tokens: %d\n", finalResponse.PromptEvalCount)
	fmt.Printf("Output tokens: %d\n", finalResponse.EvalCount)
	
	// Calculate expected cost based on default pricing
	// tinyllama: $0.00001 per 1K input, $0.00002 per 1K output
	inputCost := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
	outputCost := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
	totalCost := inputCost + outputCost
	
	fmt.Printf("Estimated cost: $%.8f (input: $%.8f, output: $%.8f)\n", 
		totalCost, inputCost, outputCost)
}

func testStreamingWithCost(ctx context.Context, client *api.Client) {
	req := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Write a haiku about OpenTelemetry",
		// Stream: nil means streaming by default
	}

	chunkCount := 0
	var lastEvalCount int
	var accumulatedCost float64

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		
		// Simulate real-time cost accumulation
		if resp.EvalCount > lastEvalCount {
			incrementalTokens := resp.EvalCount - lastEvalCount
			incrementalCost := float64(incrementalTokens) / 1000.0 * 0.00002
			accumulatedCost += incrementalCost
			
			if chunkCount%10 == 0 {
				fmt.Printf("  Chunk %d: %d tokens, accumulated cost: $%.8f\n", 
					chunkCount, resp.EvalCount, accumulatedCost)
			}
			
			lastEvalCount = resp.EvalCount
		}
		
		if resp.Done {
			fmt.Printf("\nStreaming completed:\n")
			fmt.Printf("Total chunks: %d\n", chunkCount)
			fmt.Printf("Input tokens: %d\n", resp.PromptEvalCount)
			fmt.Printf("Output tokens: %d\n", resp.EvalCount)
			
			// Final cost calculation
			inputCost := float64(resp.PromptEvalCount) / 1000.0 * 0.00001
			totalCost := inputCost + accumulatedCost
			fmt.Printf("Final cost: $%.8f (input: $%.8f, output: $%.8f)\n",
				totalCost, inputCost, accumulatedCost)
		}
		
		return nil
	})

	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func testBudgetTracking(ctx context.Context, client *api.Client) {
	// Simulate multiple requests to test budget consumption
	prompts := []string{
		"What is 2+2?",
		"Name a color",
		"Say hello",
	}
	
	totalCost := 0.0
	
	for i, prompt := range prompts {
		req := &api.GenerateRequest{
			Model:  "tinyllama",
			Prompt: prompt,
			Stream: new(bool), // false
		}
		
		var finalResponse api.GenerateResponse
		err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
			if resp.Done {
				finalResponse = resp
			}
			return nil
		})
		
		if err != nil {
			log.Printf("Request %d error: %v", i+1, err)
			continue
		}
		
		// Calculate cost for this request
		inputCost := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
		outputCost := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
		requestCost := inputCost + outputCost
		totalCost += requestCost
		
		fmt.Printf("Request %d: '%s' - Cost: $%.8f\n", i+1, prompt, requestCost)
		
		// Simulate budget check (in real implementation, this would be automatic)
		budgetLimit := 0.001 // $0.001 budget
		usagePercent := (totalCost / budgetLimit) * 100
		
		status := "OK"
		if usagePercent >= 80 {
			status = "WARNING"
		}
		if usagePercent >= 90 {
			status = "CRITICAL"
		}
		if usagePercent >= 100 {
			status = "EXCEEDED"
		}
		
		fmt.Printf("  Budget status: %s (%.1f%% of $%.6f used)\n", 
			status, usagePercent, budgetLimit)
		
		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("\nTotal cost across %d requests: $%.8f\n", len(prompts), totalCost)
}

func testCurrencyConversion(ctx context.Context, client *api.Client) {
	// Test with a simple request
	req := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Hi",
		Stream: new(bool), // false
	}
	
	var finalResponse api.GenerateResponse
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			finalResponse = resp
		}
		return nil
	})
	
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	// Calculate cost in different currencies
	inputCostUSD := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
	outputCostUSD := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
	totalCostUSD := inputCostUSD + outputCostUSD
	
	// Exchange rates (from default configuration)
	rates := map[string]float64{
		"USD": 1.0,
		"EUR": 0.85,
		"CNY": 7.25,
		"GBP": 0.79,
		"JPY": 149.50,
	}
	
	fmt.Printf("Cost in different currencies:\n")
	for currency, rate := range rates {
		convertedCost := totalCostUSD * rate
		symbol := getCurrencySymbol(currency)
		fmt.Printf("  %s: %s%.8f\n", currency, symbol, convertedCost)
	}
}

func getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"CNY": "¥",
		"GBP": "£",
		"JPY": "¥",
	}
	if symbol, ok := symbols[currency]; ok {
		return symbol
	}
	return ""
}