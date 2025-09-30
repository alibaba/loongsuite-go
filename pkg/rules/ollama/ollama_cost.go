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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	CNY Currency = "CNY"
	GBP Currency = "GBP"
	JPY Currency = "JPY"
)

type BudgetPeriod string

const (
	BudgetPeriodHourly  BudgetPeriod = "hourly"
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

type BudgetStatus string

const (
	BudgetOK       BudgetStatus = "ok"
	BudgetWarning  BudgetStatus = "warning"
	BudgetCritical BudgetStatus = "critical"
	BudgetExceeded BudgetStatus = "exceeded"
)

type CostMetrics struct {
	InputTokens     int
	OutputTokens    int
	InputCost       float64
	OutputCost      float64
	TotalCost       float64
	Currency        Currency
	ModelID         string
	PricingTier     string
	EstimatedInput  bool
	Timestamp       time.Time
}

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

type CostCalculator struct {
	pricingDB        *PricingDatabase
	enableCalculation bool
	defaultCurrency  Currency
	
	totalCost        atomic.Value
	requestCount     atomic.Int64
	tokenCount       atomic.Int64
}

type ModelPricing struct {
	ModelID         string
	InputCostPer1K  float64
	OutputCostPer1K float64
	Currency        Currency
	Tier            string
	CurrencyRates   map[Currency]float64
}

type PricingDatabase struct {
	mu       sync.RWMutex
	prices   map[string]*ModelPricing
	currency Currency
	rates    map[Currency]float64
}

type BudgetThreshold struct {
	Percentage float64
	Status     BudgetStatus
	Action     string
}

type BudgetConfig struct {
	TotalBudget   float64
	Currency      Currency
	Period        BudgetPeriod
	Thresholds    []BudgetThreshold
	WindowSize    time.Duration
	AllowOverage  bool
	ErrorBudget   float64
}

type BudgetTracker struct {
	mu            sync.RWMutex
	config        *BudgetConfig
	currentSpend  float64
	startTime     time.Time
	lastReset     time.Time
	currentStatus BudgetStatus
	
	slidingWindow []costDataPoint
	windowStart   time.Time
	
	costHistory   []float64
	movingAverage float64
	stdDeviation  float64
	
	anomalyCount  int
	lastAnomaly   *time.Time
}

type costDataPoint struct {
	timestamp time.Time
	cost      float64
}

var (
	costCalculator *CostCalculator
	pricingDB *PricingDatabase
	budgetTracker   *BudgetTracker
)

var defaultPricing = map[string]*ModelPricing{
	"tinyllama": {
		ModelID:         "tinyllama",
		InputCostPer1K:  0.00001,
		OutputCostPer1K: 0.00002,
		Currency:        USD,
		Tier:            "economy",
	},
	"llama3:8b": {
		ModelID:         "llama3:8b",
		InputCostPer1K:  0.00005,
		OutputCostPer1K: 0.0001,
		Currency:        USD,
		Tier:            "standard",
	},
	"llama3:70b": {
		ModelID:         "llama3:70b",
		InputCostPer1K:  0.0002,
		OutputCostPer1K: 0.0004,
		Currency:        USD,
		Tier:            "premium",
	},
	"mistral:7b": {
		ModelID:         "mistral:7b",
		InputCostPer1K:  0.00004,
		OutputCostPer1K: 0.00008,
		Currency:        USD,
		Tier:            "standard",
	},
	"codellama:13b": {
		ModelID:         "codellama:13b",
		InputCostPer1K:  0.00008,
		OutputCostPer1K: 0.00016,
		Currency:        USD,
		Tier:            "standard",
	},
	"gemma:2b": {
		ModelID:         "gemma:2b",
		InputCostPer1K:  0.00002,
		OutputCostPer1K: 0.00004,
		Currency:        USD,
		Tier:            "economy",
	},
	"qwen:7b": {
		ModelID:         "qwen:7b",
		InputCostPer1K:  0.00004,
		OutputCostPer1K: 0.00008,
		Currency:        USD,
		Tier:            "standard",
	},
}

var defaultExchangeRates = map[Currency]float64{
	USD: 1.0,
	EUR: 0.85,
	CNY: 7.25,
	GBP: 0.73,
	JPY: 149.5,
}

func init() {
	pricingDB = &PricingDatabase{
		prices:   make(map[string]*ModelPricing),
		currency: USD,
		rates:    defaultExchangeRates,
	}
	
	for modelID, pricing := range defaultPricing {
		pricingDB.prices[modelID] = pricing
		baseName := strings.Split(modelID, ":")[0]
		if baseName != modelID {
			pricingDB.prices[baseName] = pricing
		}
	}
	
	pricingDB.loadCustomPricing()

	enabledStr := "true"
	if val := os.Getenv("OLLAMA_ENABLE_COST_TRACKING"); val != "" {
		enabledStr = val
	}
	enabled := enabledStr != "false" && enabledStr != "0"
	
	currencyStr := "USD"
	if val := os.Getenv("OLLAMA_DEFAULT_CURRENCY"); val != "" {
		currencyStr = val
	}
	
	costCalculator = &CostCalculator{
		pricingDB:        pricingDB,
		enableCalculation: enabled,
		defaultCurrency:  Currency(currencyStr),
	}
	costCalculator.totalCost.Store(float64(0))

	config := getDefaultBudgetConfig()
	
	budgetTracker = &BudgetTracker{
		config:        config,
		currentSpend:  0,
		startTime:     time.Now(),
		lastReset:     time.Now(),
		currentStatus: BudgetOK,
		slidingWindow: make([]costDataPoint, 0),
		windowStart:   time.Now(),
		costHistory:   make([]float64, 0, 100),
	}
	
	if config.Period != "" {
		go budgetTracker.startPeriodicReset(config.Period)
	}
}


func (c *CostCalculator) IsEnabled() bool {
	return c.enableCalculation
}

func (c *CostCalculator) SetEnabled(enabled bool) {
	c.enableCalculation = enabled
}

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
	
	inputCost := float64(inputTokens) / 1000.0 * pricing.InputCostPer1K
	outputCost := float64(outputTokens) / 1000.0 * pricing.OutputCostPer1K
	
	targetCurrency := c.defaultCurrency
	if targetCurrency != pricing.Currency {
		inputCost = c.pricingDB.ConvertCurrency(inputCost, pricing.Currency, targetCurrency)
		outputCost = c.pricingDB.ConvertCurrency(outputCost, pricing.Currency, targetCurrency)
	}
	
	totalCost := inputCost + outputCost
	
	c.updateGlobalMetrics(totalCost, int64(inputTokens+outputTokens))
	
	return &CostMetrics{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputCost:    math.Round(inputCost*1000000) / 1000000,
		OutputCost:   math.Round(outputCost*1000000) / 1000000,
		TotalCost:    math.Round(totalCost*1000000) / 1000000,
		Currency:     targetCurrency,
		ModelID:      modelID,
		PricingTier:  pricing.Tier,
		Timestamp:    time.Now(),
	}, nil
}

func (c *CostCalculator) EstimateInputTokens(prompt string) int {
	if prompt == "" {
		return 0
	}
	
	estimatedTokens := len(prompt) / 4
	
	if estimatedTokens == 0 {
		estimatedTokens = 1
	}
	
	return estimatedTokens
}

func (c *CostCalculator) PredictCost(modelID string, estimatedInput, estimatedOutput int) (*CostMetrics, error) {
	metrics, err := c.CalculateCost(modelID, estimatedInput, estimatedOutput)
	if metrics != nil {
		metrics.EstimatedInput = true
	}
	return metrics, err
}

func (c *CostCalculator) NewStreamingCostState(modelID string) *StreamingCostState {
	pricing, _ := c.pricingDB.GetModelPricing(modelID)
	return &StreamingCostState{
		modelID:  modelID,
		currency: c.defaultCurrency,
		pricing:  pricing,
	}
}

func (s *StreamingCostState) UpdateStreamingCost(newTotalTokens int) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.pricing == nil {
		return 0.0
	}
	
	incrementalTokens := newTotalTokens - s.lastUpdateTokens
	if incrementalTokens <= 0 {
		return s.accumulatedCost
	}
	
	s.outputTokens += incrementalTokens
	s.lastUpdateTokens = newTotalTokens
	
	incrementalCost := float64(incrementalTokens) / 1000.0 * s.pricing.OutputCostPer1K
	s.accumulatedCost += incrementalCost
	
	return s.accumulatedCost
}

func (s *StreamingCostState) SetInputTokens(tokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inputTokens = tokens
}

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
		InputCost:    math.Round(inputCost*1000000) / 1000000,
		OutputCost:   math.Round(outputCost*1000000) / 1000000,
		TotalCost:    math.Round((inputCost + outputCost)*1000000) / 1000000,
		Currency:     s.currency,
		ModelID:      s.modelID,
		PricingTier:  func() string { if s.pricing == nil { return "unknown" }; if s.pricing.Tier == "" { return "standard" }; return s.pricing.Tier }(),
		Timestamp:    time.Now(),
	}
}

func (c *CostCalculator) updateGlobalMetrics(cost float64, tokens int64) {
	for {
		oldVal := c.totalCost.Load().(float64)
		newVal := oldVal + cost
		if c.totalCost.CompareAndSwap(oldVal, newVal) {
			break
		}
	}
	
	c.requestCount.Add(1)
	c.tokenCount.Add(tokens)
}

func (c *CostCalculator) GetGlobalMetrics() map[string]interface{} {
	totalCost := c.totalCost.Load().(float64)
	return map[string]interface{}{
		"total_cost":     math.Round(totalCost*1000000) / 1000000,
		"request_count":  c.requestCount.Load(),
		"token_count":    c.tokenCount.Load(),
		"currency":       string(c.defaultCurrency),
		"cost_per_request": math.Round(totalCost / float64(func() int64 { count := c.requestCount.Load(); if count < 1 { return 1 }; return count }()) * 1000000) / 1000000,
	}
}

func (c *CostCalculator) ResetGlobalMetrics() {
	c.totalCost.Store(float64(0))
	c.requestCount.Store(0)
	c.tokenCount.Store(0)
}



func (db *PricingDatabase) loadCustomPricing() {
	if configPath := os.Getenv("OLLAMA_COST_CONFIG"); configPath != "" {
		if err := db.LoadFromFile(configPath); err != nil {
			fmt.Printf("Warning: Failed to load custom pricing from %s: %v\n", configPath, err)
		} else {
			return
		}
	}
	
	configPaths := []string{
		"./ollama_cost_config.json",
		"~/.ollama/cost_config.json",
		"/etc/ollama/cost_config.json",
	}
	
	for _, path := range configPaths {
		if err := db.LoadFromFile(path); err == nil {
			return
		}
	}
}

func (db *PricingDatabase) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	
	var config struct {
		DefaultCurrency string                      `json:"default_currency,omitempty"`
		ExchangeRates   map[Currency]float64        `json:"exchange_rates,omitempty"`
		ModelPricing    map[string]*ModelPricing    `json:"model_pricing,omitempty"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	
	db.mu.Lock()
	defer db.mu.Unlock()
	
	if config.DefaultCurrency != "" {
		db.currency = Currency(config.DefaultCurrency)
	}
	
	if len(config.ExchangeRates) > 0 {
		db.rates = config.ExchangeRates
	}
	
	for modelID, pricing := range config.ModelPricing {
		pricing.ModelID = modelID
		db.prices[modelID] = pricing
	}
	
	return nil
}

func (db *PricingDatabase) GetModelPricing(modelID string) (*ModelPricing, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if pricing, exists := db.prices[modelID]; exists {
		return pricing, true
	}
	
	baseName := strings.Split(modelID, ":")[0]
	if pricing, exists := db.prices[baseName]; exists {
		return pricing, true
	}
	
	lowerID := strings.ToLower(modelID)
	if pricing, exists := db.prices[lowerID]; exists {
		return pricing, true
	}
	
	return nil, false
}

func (db *PricingDatabase) SetModelPricing(modelID string, pricing *ModelPricing) {
	db.mu.Lock()
	defer db.mu.Unlock()
	pricing.ModelID = modelID
	db.prices[modelID] = pricing
}

func (db *PricingDatabase) GetCurrency() Currency {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.currency
}

func (db *PricingDatabase) SetCurrency(currency Currency) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.currency = currency
}

func (db *PricingDatabase) ConvertCurrency(amount float64, from, to Currency) float64 {
	if from == to {
		return amount
	}
	
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	amountInUSD := amount
	if from != USD {
		if fromRate, exists := db.rates[from]; exists {
			amountInUSD = amount / fromRate
		}
	}
	
	if to != USD {
		if toRate, exists := db.rates[to]; exists {
			return amountInUSD * toRate
		}
	}
	
	return amountInUSD
}

func (db *PricingDatabase) GetExchangeRate(currency Currency) float64 {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if rate, exists := db.rates[currency]; exists {
		return rate
	}
	return 1.0
}

func (db *PricingDatabase) ListModels() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	models := make([]string, 0, len(db.prices))
	for modelID := range db.prices {
		models = append(models, modelID)
	}
	return models
}

func (db *PricingDatabase) ExportConfig() ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	config := struct {
		DefaultCurrency string                   `json:"default_currency"`
		ExchangeRates   map[Currency]float64     `json:"exchange_rates"`
		ModelPricing    map[string]*ModelPricing `json:"model_pricing"`
	}{
		DefaultCurrency: string(db.currency),
		ExchangeRates:   db.rates,
		ModelPricing:    db.prices,
	}
	
	return json.MarshalIndent(config, "", "  ")
}

var defaultThresholds = []BudgetThreshold{
	{Percentage: 80, Status: BudgetWarning, Action: "log"},
	{Percentage: 90, Status: BudgetCritical, Action: "alert"},
	{Percentage: 100, Status: BudgetExceeded, Action: "block"},
}



func getDefaultBudgetConfig() *BudgetConfig {
	return &BudgetConfig{
		TotalBudget:  100.0,
		Currency:     USD,
		Period:       BudgetPeriodDaily,
		Thresholds:   defaultThresholds,
		WindowSize:   time.Hour,
		AllowOverage: true,
		ErrorBudget:  10.0,
	}
}

func (bt *BudgetTracker) RecordCost(cost float64) BudgetStatus {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	bt.currentSpend += cost
	
	bt.updateSlidingWindow(cost)
	
	bt.updateCostHistory(cost)
	
	if bt.isAnomaly(cost) {
		bt.anomalyCount++
		now := time.Now()
		bt.lastAnomaly = &now
	}
	
	bt.currentStatus = bt.getCurrentStatus()
	return bt.currentStatus
}

func (bt *BudgetTracker) updateSlidingWindow(cost float64) {
	now := time.Now()
	
	bt.slidingWindow = append(bt.slidingWindow, costDataPoint{
		timestamp: now,
		cost:      cost,
	})
	
	cutoff := now.Add(-bt.config.WindowSize)
	i := 0
	for i < len(bt.slidingWindow) && bt.slidingWindow[i].timestamp.Before(cutoff) {
		i++
	}
	bt.slidingWindow = bt.slidingWindow[i:]
}

func (bt *BudgetTracker) updateCostHistory(cost float64) {
	bt.costHistory = append(bt.costHistory, cost)
	
	if len(bt.costHistory) > 100 {
		bt.costHistory = bt.costHistory[len(bt.costHistory)-100:]
	}
	
	if len(bt.costHistory) > 0 {
		sum := 0.0
		for _, c := range bt.costHistory {
			sum += c
		}
		bt.movingAverage = sum / float64(len(bt.costHistory))
		
		if len(bt.costHistory) > 1 {
			variance := 0.0
			for _, c := range bt.costHistory {
				diff := c - bt.movingAverage
				variance += diff * diff
			}
			bt.stdDeviation = math.Sqrt(variance / float64(len(bt.costHistory)-1))
		}
	}
}

func (bt *BudgetTracker) isAnomaly(cost float64) bool {
	if bt.stdDeviation == 0 || len(bt.costHistory) < 10 {
		return false
	}
	
	zScore := math.Abs((cost - bt.movingAverage) / bt.stdDeviation)
	
	return zScore > 3.0
}

func (bt *BudgetTracker) getCurrentStatus() BudgetStatus {
	percentage := (bt.currentSpend / bt.config.TotalBudget) * 100
	
	for i := len(bt.config.Thresholds) - 1; i >= 0; i-- {
		if percentage >= bt.config.Thresholds[i].Percentage {
			return bt.config.Thresholds[i].Status
		}
	}
	
	return BudgetOK
}

func (bt *BudgetTracker) GetStatus() (BudgetStatus, float64, float64) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	percentage := (bt.currentSpend / bt.config.TotalBudget) * 100
	remaining := bt.config.TotalBudget - bt.currentSpend
	
	if bt.config.ErrorBudget > 0 && bt.stdDeviation > 0 {
		_ = (bt.stdDeviation / bt.movingAverage) * 100
	}
	
	return bt.currentStatus, percentage, remaining
}

func (bt *BudgetTracker) CheckThreshold(percentage float64) bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	currentPercentage := (bt.currentSpend / bt.config.TotalBudget) * 100
	return currentPercentage >= percentage
}

func (bt *BudgetTracker) GetRemainingBudget() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	return bt.config.TotalBudget - bt.currentSpend
}

func (bt *BudgetTracker) ResetBudget() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	bt.currentSpend = 0
	bt.lastReset = time.Now()
	bt.currentStatus = BudgetOK
	bt.anomalyCount = 0
	bt.lastAnomaly = nil
	
	if len(bt.costHistory) > 10 {
		bt.costHistory = bt.costHistory[len(bt.costHistory)-10:]
	}
}

func (bt *BudgetTracker) startPeriodicReset(period BudgetPeriod) {
	ticker := bt.getResetTicker(period)
	for range ticker.C {
		bt.ResetBudget()
	}
}

func (bt *BudgetTracker) getResetTicker(period BudgetPeriod) *time.Ticker {
	switch period {
	case BudgetPeriodHourly:
		return time.NewTicker(time.Hour)
	case BudgetPeriodDaily:
		return time.NewTicker(24 * time.Hour)
	case BudgetPeriodWeekly:
		return time.NewTicker(7 * 24 * time.Hour)
	case BudgetPeriodMonthly:
		return time.NewTicker(30 * 24 * time.Hour)
	default:
		return time.NewTicker(24 * time.Hour)
	}
}

func (bt *BudgetTracker) PredictBudgetExhaustion() (*time.Time, float64) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	if len(bt.slidingWindow) < 2 {
		return nil, 0
	}
	
	windowDuration := bt.config.WindowSize
	windowCost := 0.0
	for _, dp := range bt.slidingWindow {
		windowCost += dp.cost
	}
	
	burnRate := windowCost / windowDuration.Hours()
	
	if burnRate <= 0 {
		return nil, 0
	}
	
	remainingBudget := bt.config.TotalBudget - bt.currentSpend
	hoursUntilExhaustion := remainingBudget / burnRate
	
	exhaustionTime := time.Now().Add(time.Duration(hoursUntilExhaustion) * time.Hour)
	return &exhaustionTime, burnRate
}

func (bt *BudgetTracker) GetAnomalyReport() map[string]interface{} {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	report := map[string]interface{}{
		"anomaly_count":   bt.anomalyCount,
		"moving_average":  bt.movingAverage,
		"std_deviation":   bt.stdDeviation,
		"last_anomaly":    bt.lastAnomaly,
	}
	
	if exhaustionTime, burnRate := bt.PredictBudgetExhaustion(); exhaustionTime != nil {
		report["predicted_exhaustion"] = exhaustionTime
		report["burn_rate_per_hour"] = burnRate
	}
	
	return report
}

func (bt *BudgetTracker) UpdateConfig(config *BudgetConfig) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.config = config
}

type SLOConfig struct {
	LatencyThreshold    time.Duration
	ErrorRateThreshold  float64
	P50Target           time.Duration
	P95Target           time.Duration
	P99Target           time.Duration
	WindowSize          time.Duration
	EvaluationInterval  time.Duration
}

type SLOTracker struct {
	config           *SLOConfig
	latencies        []time.Duration
	errors           []bool
	mu               sync.RWMutex
	lastEvaluation   time.Time
	p50              time.Duration
	p95              time.Duration
	p99              time.Duration
	errorRate        float64
	violations       int
	totalRequests    int64
}

var sloTracker *SLOTracker

func init() {
	sloTracker = &SLOTracker{
		config: &SLOConfig{
			LatencyThreshold:   2 * time.Second,
			ErrorRateThreshold: 0.01,
			P50Target:          500 * time.Millisecond,
			P95Target:          1500 * time.Millisecond,
			P99Target:          3000 * time.Millisecond,
			WindowSize:         5 * time.Minute,
			EvaluationInterval: 1 * time.Minute,
		},
		latencies:      make([]time.Duration, 0, 1000),
		errors:         make([]bool, 0, 1000),
		lastEvaluation: time.Now(),
	}
}

func (st *SLOTracker) RecordRequest(latency time.Duration, isError bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.latencies = append(st.latencies, latency)
	st.errors = append(st.errors, isError)
	atomic.AddInt64(&st.totalRequests, 1)

	if len(st.latencies) > 10000 {
		st.latencies = st.latencies[5000:]
		st.errors = st.errors[5000:]
	}

	if time.Since(st.lastEvaluation) > st.config.EvaluationInterval {
		st.evaluate()
	}
}

func (st *SLOTracker) evaluate() {
	if len(st.latencies) == 0 {
		return
	}

	sortedLatencies := make([]time.Duration, len(st.latencies))
	copy(sortedLatencies, st.latencies)

	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	st.p50 = sortedLatencies[len(sortedLatencies)*50/100]
	st.p95 = sortedLatencies[len(sortedLatencies)*95/100]
	st.p99 = sortedLatencies[len(sortedLatencies)*99/100]

	errorCount := 0
	for _, e := range st.errors {
		if e {
			errorCount++
		}
	}
	st.errorRate = float64(errorCount) / float64(len(st.errors))

	if st.p50 > st.config.P50Target || st.p95 > st.config.P95Target ||
	   st.p99 > st.config.P99Target || st.errorRate > st.config.ErrorRateThreshold {
		st.violations++
	}

	st.lastEvaluation = time.Now()
}

func (st *SLOTracker) GetMetrics() map[string]interface{} {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return map[string]interface{}{
		"p50_ms":           st.p50.Milliseconds(),
		"p95_ms":           st.p95.Milliseconds(),
		"p99_ms":           st.p99.Milliseconds(),
		"error_rate":       st.errorRate,
		"violations":       st.violations,
		"total_requests":   atomic.LoadInt64(&st.totalRequests),
		"slo_compliance":   st.getCompliance(),
	}
}

func (st *SLOTracker) getCompliance() float64 {
	if st.totalRequests == 0 {
		return 100.0
	}
	return (1.0 - float64(st.violations)/float64(st.totalRequests)) * 100.0
}

func (st *SLOTracker) IsPerformanceBottleneck(latency time.Duration) bool {
	return latency > st.config.LatencyThreshold
}

func (st *SLOTracker) DetectQualityDegradation() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if len(st.errors) < 100 {
		return false
	}

	recentErrors := st.errors[len(st.errors)-100:]
	recentErrorCount := 0
	for _, e := range recentErrors {
		if e {
			recentErrorCount++
		}
	}

	recentErrorRate := float64(recentErrorCount) / 100.0
	return recentErrorRate > st.config.ErrorRateThreshold * 2
}

func calculateEmbeddingCost(modelID string, embeddingCount int, dimensions int) *CostMetrics {
	calculator := costCalculator
	if calculator == nil || !calculator.IsEnabled() {
		return nil
	}

	estimatedTokens := embeddingCount * (dimensions / 4)

	metrics, _ := calculator.CalculateCost(modelID, estimatedTokens, 0)
	if metrics != nil {
		metrics.PricingTier = "embedding"
	}
	return metrics
}