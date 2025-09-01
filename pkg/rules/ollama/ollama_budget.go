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
	"sync"
	"time"
)

// BudgetPeriod represents the budget reset period
type BudgetPeriod string

const (
	Hourly  BudgetPeriod = "hourly"
	Daily   BudgetPeriod = "daily"
	Weekly  BudgetPeriod = "weekly"
	Monthly BudgetPeriod = "monthly"
	NoReset BudgetPeriod = "no_reset"
)

// BudgetStatus represents the current budget status
type BudgetStatus string

const (
	BudgetOK       BudgetStatus = "ok"
	BudgetWarning  BudgetStatus = "warning"
	BudgetCritical BudgetStatus = "critical"
	BudgetExceeded BudgetStatus = "exceeded"
)

// BudgetThreshold represents a budget threshold level
type BudgetThreshold struct {
	Percentage float64      `json:"percentage"`
	Status     BudgetStatus `json:"status"`
	Action     string       `json:"action"` // e.g., "log", "alert", "block"
}

// BudgetConfig represents budget configuration
type BudgetConfig struct {
	TotalBudget   float64          `json:"total_budget"`
	Currency      Currency         `json:"currency"`
	Period        BudgetPeriod     `json:"period"`
	Thresholds    []BudgetThreshold `json:"thresholds"`
	WindowSize    time.Duration    `json:"window_size"`    // For sliding window
	AllowOverage  bool             `json:"allow_overage"`  // Whether to allow exceeding budget
	ErrorBudget   float64          `json:"error_budget"`   // Acceptable variance percentage
}

// BudgetTracker tracks cost budgets with SRE-style error budgets
type BudgetTracker struct {
	mu            sync.RWMutex
	config        *BudgetConfig
	currentSpend  float64
	periodStart   time.Time
	lastReset     time.Time
	
	// Sliding window tracking
	windowData    []costDataPoint
	windowSum     float64
	
	// Cost variance tracking (for error budget)
	costHistory   []float64
	movingAverage float64
	stdDeviation  float64
	
	// Anomaly detection
	anomalyCount  int
	lastAnomaly   time.Time
}

// costDataPoint represents a cost data point in time
type costDataPoint struct {
	timestamp time.Time
	cost      float64
}

var (
	// Global budget tracker instance
	globalBudget *BudgetTracker
	budgetOnce   sync.Once
	
	// Default thresholds
	defaultThresholds = []BudgetThreshold{
		{Percentage: 80, Status: BudgetWarning, Action: "log"},
		{Percentage: 90, Status: BudgetCritical, Action: "alert"},
		{Percentage: 100, Status: BudgetExceeded, Action: "alert"},
	}
)

// InitializeBudgetTracker initializes the global budget tracker
func InitializeBudgetTracker(config *BudgetConfig) *BudgetTracker {
	budgetOnce.Do(func() {
		if config == nil {
			config = getDefaultBudgetConfig()
		}
		
		globalBudget = &BudgetTracker{
			config:       config,
			periodStart:  time.Now(),
			lastReset:    time.Now(),
			windowData:   make([]costDataPoint, 0),
			costHistory:  make([]float64, 0, 100),
		}
		
		// Start periodic reset if needed
		if config.Period != NoReset {
			go globalBudget.startPeriodicReset()
		}
	})
	return globalBudget
}

// GetBudgetTracker returns the global budget tracker instance
func GetBudgetTracker() *BudgetTracker {
	if globalBudget == nil {
		return InitializeBudgetTracker(nil)
	}
	return globalBudget
}

// getDefaultBudgetConfig returns default budget configuration
func getDefaultBudgetConfig() *BudgetConfig {
	return &BudgetConfig{
		TotalBudget:  100.0, // $100 USD default
		Currency:     USD,
		Period:       Daily,
		Thresholds:   defaultThresholds,
		WindowSize:   time.Hour,
		AllowOverage: true,
		ErrorBudget:  10.0, // 10% variance allowed
	}
}

// RecordCost records a cost and checks budget status
func (bt *BudgetTracker) RecordCost(cost float64) BudgetStatus {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	// Update current spend
	bt.currentSpend += cost
	
	// Update sliding window
	bt.updateSlidingWindow(cost)
	
	// Update cost history for variance tracking
	bt.updateCostHistory(cost)
	
	// Check for anomalies
	if bt.isAnomaly(cost) {
		bt.anomalyCount++
		bt.lastAnomaly = time.Now()
	}
	
	// Determine current status
	return bt.getCurrentStatus()
}

// updateSlidingWindow updates the sliding window data
func (bt *BudgetTracker) updateSlidingWindow(cost float64) {
	now := time.Now()
	
	// Add new data point
	bt.windowData = append(bt.windowData, costDataPoint{
		timestamp: now,
		cost:      cost,
	})
	
	// Remove old data points outside window
	cutoff := now.Add(-bt.config.WindowSize)
	newWindowData := make([]costDataPoint, 0)
	newWindowSum := 0.0
	
	for _, dp := range bt.windowData {
		if dp.timestamp.After(cutoff) {
			newWindowData = append(newWindowData, dp)
			newWindowSum += dp.cost
		}
	}
	
	bt.windowData = newWindowData
	bt.windowSum = newWindowSum
}

// updateCostHistory updates cost history for variance tracking
func (bt *BudgetTracker) updateCostHistory(cost float64) {
	bt.costHistory = append(bt.costHistory, cost)
	
	// Keep only last 100 entries for efficiency
	if len(bt.costHistory) > 100 {
		bt.costHistory = bt.costHistory[len(bt.costHistory)-100:]
	}
	
	// Recalculate moving average and standard deviation
	if len(bt.costHistory) >= 10 {
		sum := 0.0
		for _, c := range bt.costHistory {
			sum += c
		}
		bt.movingAverage = sum / float64(len(bt.costHistory))
		
		// Calculate standard deviation
		variance := 0.0
		for _, c := range bt.costHistory {
			diff := c - bt.movingAverage
			variance += diff * diff
		}
		bt.stdDeviation = math.Sqrt(variance / float64(len(bt.costHistory)))
	}
}

// isAnomaly detects if a cost is anomalous using z-score
func (bt *BudgetTracker) isAnomaly(cost float64) bool {
	if bt.stdDeviation == 0 || len(bt.costHistory) < 10 {
		return false
	}
	
	// Calculate z-score
	zScore := math.Abs(cost-bt.movingAverage) / bt.stdDeviation
	
	// Anomaly if z-score > 3 (99.7% confidence)
	return zScore > 3.0
}

// getCurrentStatus determines the current budget status
func (bt *BudgetTracker) getCurrentStatus() BudgetStatus {
	if bt.config.TotalBudget <= 0 {
		return BudgetOK
	}
	
	percentage := (bt.currentSpend / bt.config.TotalBudget) * 100
	
	// Check thresholds in reverse order (highest first)
	for i := len(bt.config.Thresholds) - 1; i >= 0; i-- {
		threshold := bt.config.Thresholds[i]
		if percentage >= threshold.Percentage {
			return threshold.Status
		}
	}
	
	return BudgetOK
}

// GetStatus returns the current budget status with details
func (bt *BudgetTracker) GetStatus() map[string]interface{} {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	status := bt.getCurrentStatus()
	percentage := 0.0
	if bt.config.TotalBudget > 0 {
		percentage = (bt.currentSpend / bt.config.TotalBudget) * 100
	}
	
	// Calculate error budget consumption
	errorBudgetConsumed := 0.0
	if bt.config.ErrorBudget > 0 && bt.stdDeviation > 0 {
		// Error budget consumed = actual variance / allowed variance
		actualVariance := (bt.stdDeviation / bt.movingAverage) * 100
		errorBudgetConsumed = (actualVariance / bt.config.ErrorBudget) * 100
	}
	
	return map[string]interface{}{
		"status":                string(status),
		"current_spend":         roundToMicros(bt.currentSpend),
		"total_budget":          bt.config.TotalBudget,
		"percentage_used":       roundToMicros(percentage),
		"currency":              string(bt.config.Currency),
		"period":                string(bt.config.Period),
		"period_start":          bt.periodStart,
		"window_spend":          roundToMicros(bt.windowSum),
		"window_size":           bt.config.WindowSize.String(),
		"anomaly_count":         bt.anomalyCount,
		"last_anomaly":          bt.lastAnomaly,
		"moving_average":        roundToMicros(bt.movingAverage),
		"std_deviation":         roundToMicros(bt.stdDeviation),
		"error_budget_consumed": roundToMicros(errorBudgetConsumed),
	}
}

// CheckThreshold checks if a specific threshold has been exceeded
func (bt *BudgetTracker) CheckThreshold(percentage float64) bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	if bt.config.TotalBudget <= 0 {
		return false
	}
	
	currentPercentage := (bt.currentSpend / bt.config.TotalBudget) * 100
	return currentPercentage >= percentage
}

// GetRemainingBudget returns the remaining budget
func (bt *BudgetTracker) GetRemainingBudget() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	remaining := bt.config.TotalBudget - bt.currentSpend
	if remaining < 0 && !bt.config.AllowOverage {
		return 0
	}
	return remaining
}

// ResetBudget resets the budget for a new period
func (bt *BudgetTracker) ResetBudget() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	
	bt.currentSpend = 0
	bt.periodStart = time.Now()
	bt.lastReset = time.Now()
	bt.anomalyCount = 0
	
	// Keep some history for continuity
	if len(bt.costHistory) > 20 {
		bt.costHistory = bt.costHistory[len(bt.costHistory)-20:]
	}
}

// startPeriodicReset starts a goroutine for periodic budget reset
func (bt *BudgetTracker) startPeriodicReset() {
	ticker := bt.getResetTicker()
	if ticker == nil {
		return
	}
	
	for range ticker.C {
		bt.ResetBudget()
	}
}

// getResetTicker returns a ticker based on the budget period
func (bt *BudgetTracker) getResetTicker() *time.Ticker {
	switch bt.config.Period {
	case Hourly:
		return time.NewTicker(time.Hour)
	case Daily:
		return time.NewTicker(24 * time.Hour)
	case Weekly:
		return time.NewTicker(7 * 24 * time.Hour)
	case Monthly:
		return time.NewTicker(30 * 24 * time.Hour)
	default:
		return nil
	}
}

// PredictBudgetExhaustion predicts when budget will be exhausted based on current rate
func (bt *BudgetTracker) PredictBudgetExhaustion() (time.Time, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	if bt.config.TotalBudget <= 0 || bt.currentSpend >= bt.config.TotalBudget {
		return time.Time{}, false
	}
	
	// Calculate burn rate (cost per hour)
	elapsed := time.Since(bt.periodStart).Hours()
	if elapsed <= 0 {
		return time.Time{}, false
	}
	
	burnRate := bt.currentSpend / elapsed
	if burnRate <= 0 {
		return time.Time{}, false
	}
	
	// Calculate hours until exhaustion
	remaining := bt.config.TotalBudget - bt.currentSpend
	hoursUntilExhaustion := remaining / burnRate
	
	// Predict exhaustion time
	exhaustionTime := time.Now().Add(time.Duration(hoursUntilExhaustion) * time.Hour)
	return exhaustionTime, true
}

// GetAnomalyReport returns a report of detected anomalies
func (bt *BudgetTracker) GetAnomalyReport() map[string]interface{} {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	
	anomalyRate := 0.0
	if len(bt.costHistory) > 0 {
		anomalyRate = float64(bt.anomalyCount) / float64(len(bt.costHistory)) * 100
	}
	
	return map[string]interface{}{
		"anomaly_count":    bt.anomalyCount,
		"anomaly_rate":     roundToMicros(anomalyRate),
		"last_anomaly":     bt.lastAnomaly,
		"threshold_zscore": 3.0,
		"current_std_dev":  roundToMicros(bt.stdDeviation),
		"current_avg":      roundToMicros(bt.movingAverage),
	}
}

// UpdateConfig updates the budget configuration
func (bt *BudgetTracker) UpdateConfig(config *BudgetConfig) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.config = config
}