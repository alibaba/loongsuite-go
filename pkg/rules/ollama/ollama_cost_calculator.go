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

package ollama

import (
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// CostMetrics represents cost-related metrics for a request
type CostMetrics struct {
	InputTokens     int     `json:"input_tokens"`
	OutputTokens    int     `json:"output_tokens"`
	InputCost       float64 `json:"input_cost"`
	OutputCost      float64 `json:"output_cost"`
	TotalCost       float64 `json:"total_cost"`
	Currency        Currency `json:"currency"`
	ModelID         string  `json:"model_id"`
	PricingTier     string  `json:"pricing_tier"`
	EstimatedInput  bool    `json:"estimated_input"` // True if input tokens were estimated
	Timestamp       time.Time `json:"timestamp"`
}

// StreamingCostState tracks cost accumulation during streaming
type StreamingCostState struct {
	mu               sync.Mutex
	modelID          string
	currency         Currency
	accumulatedCost  float64
	inputTokens      int
	outputTokens     int
	lastUpdateTokens int
	pricing          *ModelPricing
}

// CostCalculator handles cost calculation for Ollama API calls
type CostCalculator struct {
	pricingDB        *PricingDatabase
	enableCalculation bool
	defaultCurrency  Currency
	
	// Metrics tracking
	totalCost        atomic.Value // float64
	requestCount     atomic.Int64
	tokenCount       atomic.Int64
}

var (
	// Global cost calculator instance
	globalCalculator *CostCalculator
	calcInitOnce     sync.Once
)

// InitializeCostCalculator initializes the global cost calculator
func InitializeCostCalculator() *CostCalculator {
	calcInitOnce.Do(func() {
		// Check if cost tracking is disabled
		enabledStr := getEnvWithDefault("OLLAMA_ENABLE_COST_TRACKING", "true")
		enabled := enabledStr != "false" && enabledStr != "0"
		
		// Get default currency
		currencyStr := getEnvWithDefault("OLLAMA_DEFAULT_CURRENCY", "USD")
		
		globalCalculator = &CostCalculator{
			pricingDB:        InitializePricingDatabase(),
			enableCalculation: enabled,
			defaultCurrency:  Currency(currencyStr),
		}
		globalCalculator.totalCost.Store(float64(0))
	})
	return globalCalculator
}

// GetCostCalculator returns the global cost calculator instance
func GetCostCalculator() *CostCalculator {
	if globalCalculator == nil {
		return InitializeCostCalculator()
	}
	return globalCalculator
}

// IsEnabled returns whether cost calculation is enabled
func (c *CostCalculator) IsEnabled() bool {
	return c.enableCalculation
}

// SetEnabled enables or disables cost calculation
func (c *CostCalculator) SetEnabled(enabled bool) {
	c.enableCalculation = enabled
}

// CalculateCost calculates the cost for a given number of tokens
func (c *CostCalculator) CalculateCost(modelID string, inputTokens, outputTokens int) (*CostMetrics, error) {
	if !c.enableCalculation {
		return &CostMetrics{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			Currency:     c.defaultCurrency,
			ModelID:      modelID,
			Timestamp:    time.Now(),
		}, nil
	}
	
	pricing, exists := c.pricingDB.GetModelPricing(modelID)
	if !exists {
		// Return zero cost if pricing not found
		return &CostMetrics{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			InputCost:    0.0,
			OutputCost:   0.0,
			TotalCost:    0.0,
			Currency:     c.defaultCurrency,
			ModelID:      modelID,
			PricingTier:  "unknown",
			Timestamp:    time.Now(),
		}, nil
	}
	
	// Calculate costs (pricing is per 1K tokens)
	inputCost := float64(inputTokens) / 1000.0 * pricing.InputCostPer1K
	outputCost := float64(outputTokens) / 1000.0 * pricing.OutputCostPer1K
	
	// Convert to target currency if needed
	targetCurrency := c.defaultCurrency
	if targetCurrency != pricing.Currency {
		inputCost = c.pricingDB.ConvertCurrency(inputCost, pricing.Currency, targetCurrency)
		outputCost = c.pricingDB.ConvertCurrency(outputCost, pricing.Currency, targetCurrency)
	}
	
	totalCost := inputCost + outputCost
	
	// Update global metrics
	c.updateGlobalMetrics(totalCost, int64(inputTokens+outputTokens))
	
	return &CostMetrics{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputCost:    roundToMicros(inputCost),
		OutputCost:   roundToMicros(outputCost),
		TotalCost:    roundToMicros(totalCost),
		Currency:     targetCurrency,
		ModelID:      modelID,
		PricingTier:  pricing.Tier,
		Timestamp:    time.Now(),
	}, nil
}

// EstimateInputTokens estimates input tokens based on prompt length
// This is a simple estimation - actual tokenization may vary
func (c *CostCalculator) EstimateInputTokens(prompt string) int {
	if prompt == "" {
		return 0
	}
	
	// Simple estimation: ~4 characters per token (rough average for English)
	// This is a very rough estimate and varies by language and content
	estimatedTokens := len(prompt) / 4
	
	// Minimum 1 token for non-empty prompt
	if estimatedTokens == 0 {
		estimatedTokens = 1
	}
	
	return estimatedTokens
}

// PredictCost predicts the cost before making a request
func (c *CostCalculator) PredictCost(modelID string, estimatedInput, estimatedOutput int) (*CostMetrics, error) {
	metrics, err := c.CalculateCost(modelID, estimatedInput, estimatedOutput)
	if metrics != nil {
		metrics.EstimatedInput = true
	}
	return metrics, err
}

// NewStreamingCostState creates a new streaming cost state for real-time tracking
func (c *CostCalculator) NewStreamingCostState(modelID string) *StreamingCostState {
	pricing, _ := c.pricingDB.GetModelPricing(modelID)
	return &StreamingCostState{
		modelID:  modelID,
		currency: c.defaultCurrency,
		pricing:  pricing,
	}
}

// UpdateStreamingCost updates cost during streaming
func (s *StreamingCostState) UpdateStreamingCost(newTotalTokens int) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.pricing == nil {
		return 0.0
	}
	
	// Calculate incremental tokens (Ollama provides cumulative count)
	incrementalTokens := newTotalTokens - s.lastUpdateTokens
	if incrementalTokens <= 0 {
		return s.accumulatedCost
	}
	
	// Update output tokens
	s.outputTokens += incrementalTokens
	s.lastUpdateTokens = newTotalTokens
	
	// Calculate incremental cost
	incrementalCost := float64(incrementalTokens) / 1000.0 * s.pricing.OutputCostPer1K
	s.accumulatedCost += incrementalCost
	
	return s.accumulatedCost
}

// SetInputTokens sets the input token count (usually from PromptEvalCount)
func (s *StreamingCostState) SetInputTokens(tokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inputTokens = tokens
}

// GetMetrics returns the current cost metrics for streaming
func (s *StreamingCostState) GetMetrics() *CostMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	inputCost := 0.0
	outputCost := s.accumulatedCost
	
	if s.pricing != nil && s.inputTokens > 0 {
		inputCost = float64(s.inputTokens) / 1000.0 * s.pricing.InputCostPer1K
	}
	
	return &CostMetrics{
		InputTokens:  s.inputTokens,
		OutputTokens: s.outputTokens,
		InputCost:    roundToMicros(inputCost),
		OutputCost:   roundToMicros(outputCost),
		TotalCost:    roundToMicros(inputCost + outputCost),
		Currency:     s.currency,
		ModelID:      s.modelID,
		PricingTier:  getPricingTier(s.pricing),
		Timestamp:    time.Now(),
	}
}

// updateGlobalMetrics updates global cost and token metrics
func (c *CostCalculator) updateGlobalMetrics(cost float64, tokens int64) {
	// Update total cost atomically
	for {
		oldVal := c.totalCost.Load().(float64)
		newVal := oldVal + cost
		if c.totalCost.CompareAndSwap(oldVal, newVal) {
			break
		}
	}
	
	// Update counters
	c.requestCount.Add(1)
	c.tokenCount.Add(tokens)
}

// GetGlobalMetrics returns global cost metrics
func (c *CostCalculator) GetGlobalMetrics() map[string]interface{} {
	totalCost := c.totalCost.Load().(float64)
	return map[string]interface{}{
		"total_cost":     roundToMicros(totalCost),
		"request_count":  c.requestCount.Load(),
		"token_count":    c.tokenCount.Load(),
		"currency":       string(c.defaultCurrency),
		"cost_per_request": roundToMicros(totalCost / float64(max(1, c.requestCount.Load()))),
	}
}

// ResetGlobalMetrics resets global metrics
func (c *CostCalculator) ResetGlobalMetrics() {
	c.totalCost.Store(float64(0))
	c.requestCount.Store(0)
	c.tokenCount.Store(0)
}

// roundToMicros rounds to 6 decimal places (microcents)
func roundToMicros(value float64) float64 {
	return math.Round(value*1000000) / 1000000
}

// getPricingTier safely gets pricing tier
func getPricingTier(pricing *ModelPricing) string {
	if pricing == nil {
		return "unknown"
	}
	if pricing.Tier == "" {
		return "standard"
	}
	return pricing.Tier
}

// getEnvWithDefault gets environment variable with default value
func getEnvWithDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}