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
	fmt.Println("Testing Ollama budget tracking instrumentation...")
	fmt.Println("==================================================")

	os.Setenv("OLLAMA_ENABLE_COST_TRACKING", "true")
	os.Setenv("OLLAMA_DEFAULT_CURRENCY", "USD")

	budgetConfig := struct {
		TotalBudget float64
		Period      string
		Thresholds  []float64
	}{
		TotalBudget: 0.0001, // $0.0001 budget (very small for testing)
		Period:      "hourly",
		Thresholds:  []float64{80, 90, 100}, // Warning at 80%, Critical at 90%, Exceeded at 100%
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	fmt.Println("\n1. Testing progressive budget consumption...")
	testProgressiveBudgetConsumption(ctx, client, budgetConfig)

	fmt.Println("\n2. Testing budget threshold alerts...")
	testBudgetThresholds(ctx, client, budgetConfig)

	fmt.Println("\n3. Testing sliding window budget tracking...")
	testSlidingWindowBudget(ctx, client)

	fmt.Println("\n4. Testing cost anomaly detection...")
	testAnomalyDetection(ctx, client)

	fmt.Println("\n==================================================")
	fmt.Println("Budget tracking tests completed!")
}

func testProgressiveBudgetConsumption(ctx context.Context, client *api.Client, config struct {
	TotalBudget float64
	Period      string
	Thresholds  []float64
}) {
	totalSpent := 0.0
	requestCount := 0

	prompts := []string{
		"Say 'a'",
		"Say 'b'",
		"Say 'c'",
		"Say 'd'",
		"Say 'e'",
	}

	for _, prompt := range prompts {
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
			log.Printf("Request error: %v", err)
			continue
		}

		requestCount++

		inputCost := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
		outputCost := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
		requestCost := inputCost + outputCost
		totalSpent += requestCost

		usagePercent := (totalSpent / config.TotalBudget) * 100
		remaining := config.TotalBudget - totalSpent

		fmt.Printf("Request %d: Cost=$%.8f, Total=$%.8f, Usage=%.1f%%, Remaining=$%.8f\n",
			requestCount, requestCost, totalSpent, usagePercent, remaining)

		if usagePercent >= 100 {
			fmt.Println("  ⚠️  BUDGET EXCEEDED - Further requests would be monitored/blocked")
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\nSummary: %d requests, Total cost: $%.8f, Budget: $%.8f\n",
		requestCount, totalSpent, config.TotalBudget)
}

func testBudgetThresholds(ctx context.Context, client *api.Client, config struct {
	TotalBudget float64
	Period      string
	Thresholds  []float64
}) {
	fmt.Println("Simulating budget threshold alerts...")

	usageLevels := []float64{50, 75, 85, 95, 105} // Percentages

	for _, level := range usageLevels {
		status := getBudgetStatus(level, config.Thresholds)
		action := getBudgetAction(level, config.Thresholds)

		fmt.Printf("  Usage: %.0f%% - Status: %s, Action: %s\n", level, status, action)

		if level >= 80 {
			fmt.Printf("    → gen_ai.budget.status: %s\n", status)
			fmt.Printf("    → gen_ai.budget.usage_percentage: %.1f\n", level)
		}
		if level >= 100 {
			fmt.Printf("    → gen_ai.budget.threshold_exceeded: true\n")
		}
	}
}

func testSlidingWindowBudget(ctx context.Context, client *api.Client) {
	fmt.Println("Testing sliding window (last hour) budget tracking...")

	windowCosts := make([]float64, 0)
	windowStart := time.Now()

	for i := 0; i < 3; i++ {
		req := &api.GenerateRequest{
			Model:  "tinyllama",
			Prompt: fmt.Sprintf("Count to %d", i+1),
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
			log.Printf("Request error: %v", err)
			continue
		}

		inputCost := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
		outputCost := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
		requestCost := inputCost + outputCost
		windowCosts = append(windowCosts, requestCost)

		elapsed := time.Since(windowStart)
		windowSum := sum(windowCosts)

		fmt.Printf("  Request %d: Cost=$%.8f, Window sum=$%.8f, Window duration=%v\n",
			i+1, requestCost, windowSum, elapsed)

		time.Sleep(100 * time.Millisecond)
	}

	windowDuration := time.Since(windowStart)
	windowTotal := sum(windowCosts)
	burnRate := windowTotal / windowDuration.Hours() // $ per hour

	fmt.Printf("\nWindow metrics:\n")
	fmt.Printf("  Total cost in window: $%.8f\n", windowTotal)
	fmt.Printf("  Window duration: %v\n", windowDuration)
	fmt.Printf("  Burn rate: $%.8f/hour\n", burnRate)

	assumedBudget := 0.001 // $0.001
	if burnRate > 0 {
		hoursUntilExhaustion := assumedBudget / burnRate
		exhaustionTime := time.Now().Add(time.Duration(hoursUntilExhaustion * float64(time.Hour)))
		fmt.Printf("  Predicted budget exhaustion: %v (%.1f hours)\n",
			exhaustionTime.Format("15:04:05"), hoursUntilExhaustion)
	}
}

func testAnomalyDetection(ctx context.Context, client *api.Client) {
	fmt.Println("Testing cost anomaly detection (z-score based)...")

	// Normal requests (establish baseline)
	costs := make([]float64, 0)
	normalPrompts := []string{"Hi", "Hello", "Test", "OK", "Yes"}

	fmt.Println("  Establishing baseline with normal requests...")
	for i, prompt := range normalPrompts {
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
			log.Printf("Request error: %v", err)
			continue
		}

		inputCost := float64(finalResponse.PromptEvalCount) / 1000.0 * 0.00001
		outputCost := float64(finalResponse.EvalCount) / 1000.0 * 0.00002
		requestCost := inputCost + outputCost
		costs = append(costs, requestCost)

		fmt.Printf("    Request %d: $%.8f\n", i+1, requestCost)
		time.Sleep(50 * time.Millisecond)
	}

	// Calculate baseline statistics
	mean, stdDev := calculateStats(costs)
	fmt.Printf("\n  Baseline: mean=$%.8f, std_dev=$%.8f\n", mean, stdDev)

	// Anomalous request (much longer prompt)
	fmt.Println("\n  Testing anomalous request...")
	anomalousReq := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Write a very long story about a robot learning to paint. Include lots of details about colors, techniques, and emotions. Make it at least 20 sentences long with vivid descriptions.",
		Stream: new(bool), // false
	}

	var anomalousResponse api.GenerateResponse
	err := client.Generate(ctx, anomalousReq, func(resp api.GenerateResponse) error {
		if resp.Done {
			anomalousResponse = resp
		}
		return nil
	})

	if err == nil {
		// Calculate anomalous cost
		inputCost := float64(anomalousResponse.PromptEvalCount) / 1000.0 * 0.00001
		outputCost := float64(anomalousResponse.EvalCount) / 1000.0 * 0.00002
		anomalousCost := inputCost + outputCost

		// Calculate z-score
		zScore := 0.0
		if stdDev > 0 {
			zScore = (anomalousCost - mean) / stdDev
		}

		isAnomaly := zScore > 3.0 // 3 standard deviations

		fmt.Printf("    Anomalous request: $%.8f\n", anomalousCost)
		fmt.Printf("    Z-score: %.2f\n", zScore)
		fmt.Printf("    Is anomaly? %v (threshold: z > 3.0)\n", isAnomaly)

		if isAnomaly {
			fmt.Println("    ⚠️  ANOMALY DETECTED - Cost spike detected!")
			fmt.Println("    → Would trigger: gen_ai.budget.anomaly_detected: true")
		}
	}
}

// Helper functions

func getBudgetStatus(usagePercent float64, thresholds []float64) string {
	if usagePercent >= thresholds[2] { // 100%
		return "EXCEEDED"
	} else if usagePercent >= thresholds[1] { // 90%
		return "CRITICAL"
	} else if usagePercent >= thresholds[0] { // 80%
		return "WARNING"
	}
	return "OK"
}

func getBudgetAction(usagePercent float64, thresholds []float64) string {
	if usagePercent >= thresholds[2] {
		return "alert + potential block"
	} else if usagePercent >= thresholds[1] {
		return "alert"
	} else if usagePercent >= thresholds[0] {
		return "log warning"
	}
	return "none"
}

func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	if len(values) > 1 {
		variance = variance / float64(len(values)-1)
		stdDev = variance
		if variance > 0 {
			// Simple square root approximation (in real code, use math.Sqrt)
			// This is a Newton-Raphson approximation
			x := variance
			for i := 0; i < 10; i++ {
				x = (x + variance/x) / 2
			}
			stdDev = x
		}
	}

	return mean, stdDev
}

