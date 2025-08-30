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
	"os"
	"strings"
	"sync"
)

// Currency represents supported currencies for cost calculation
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	CNY Currency = "CNY"
	GBP Currency = "GBP"
	JPY Currency = "JPY"
)

// ModelPricing represents the pricing for a specific model
type ModelPricing struct {
	ModelID         string             `json:"model_id"`
	InputCostPer1K  float64            `json:"input_cost_per_1k"`  // Cost per 1000 input tokens
	OutputCostPer1K float64            `json:"output_cost_per_1k"` // Cost per 1000 output tokens
	Currency        Currency           `json:"currency"`
	Tier            string             `json:"tier"` // e.g., "standard", "premium", "economy"
	CurrencyRates   map[Currency]float64 `json:"currency_rates,omitempty"` // Exchange rates
}

// PricingDatabase manages model pricing information
type PricingDatabase struct {
	mu       sync.RWMutex
	prices   map[string]*ModelPricing
	currency Currency
	rates    map[Currency]float64 // Exchange rates relative to USD
}

var (
	// Global pricing database instance
	globalPricingDB *PricingDatabase
	dbInitOnce      sync.Once
)

// Default model pricing (in USD per 1K tokens)
// Based on relative compute requirements
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
		InputCostPer1K:  0.00003,
		OutputCostPer1K: 0.00006,
		Currency:        USD,
		Tier:            "standard",
	},
}

// Default exchange rates (relative to USD)
var defaultExchangeRates = map[Currency]float64{
	USD: 1.0,
	EUR: 0.85,
	CNY: 7.25,
	GBP: 0.79,
	JPY: 149.50,
}

// InitializePricingDatabase initializes the global pricing database
func InitializePricingDatabase() *PricingDatabase {
	dbInitOnce.Do(func() {
		globalPricingDB = &PricingDatabase{
			prices:   make(map[string]*ModelPricing),
			currency: USD,
			rates:    defaultExchangeRates,
		}
		
		// Load default pricing
		for modelID, pricing := range defaultPricing {
			globalPricingDB.prices[modelID] = pricing
			// Also support model names without version suffix
			if idx := strings.Index(modelID, ":"); idx > 0 {
				baseModel := modelID[:idx]
				if _, exists := globalPricingDB.prices[baseModel]; !exists {
					globalPricingDB.prices[baseModel] = pricing
				}
			}
		}
		
		// Try to load custom pricing from environment or config file
		globalPricingDB.loadCustomPricing()
	})
	return globalPricingDB
}

// GetPricingDatabase returns the global pricing database instance
func GetPricingDatabase() *PricingDatabase {
	if globalPricingDB == nil {
		return InitializePricingDatabase()
	}
	return globalPricingDB
}

// loadCustomPricing loads custom pricing from environment or config file
func (db *PricingDatabase) loadCustomPricing() {
	// Check environment variable first
	if configPath := os.Getenv("OLLAMA_COST_CONFIG"); configPath != "" {
		if err := db.LoadFromFile(configPath); err != nil {
			// Log error but continue with defaults
			fmt.Fprintf(os.Stderr, "Failed to load custom pricing from %s: %v\n", configPath, err)
		}
		return
	}
	
	// Check default config file locations
	configPaths := []string{
		"ollama-cost-config.json",
		"/etc/ollama/cost-config.json",
		"./config/ollama-cost.json",
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if err := db.LoadFromFile(path); err == nil {
				return // Successfully loaded
			}
		}
	}
	
	// No custom config found, using defaults
}

// LoadFromFile loads pricing configuration from a JSON file
func (db *PricingDatabase) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read pricing config: %w", err)
	}
	
	var config struct {
		Currency string                    `json:"default_currency"`
		Rates    map[string]float64        `json:"exchange_rates"`
		Models   map[string]*ModelPricing  `json:"model_pricing"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse pricing config: %w", err)
	}
	
	db.mu.Lock()
	defer db.mu.Unlock()
	
	// Update currency if specified
	if config.Currency != "" {
		db.currency = Currency(config.Currency)
	}
	
	// Update exchange rates if provided
	if config.Rates != nil {
		for curr, rate := range config.Rates {
			db.rates[Currency(curr)] = rate
		}
	}
	
	// Update model pricing
	if config.Models != nil {
		for modelID, pricing := range config.Models {
			pricing.ModelID = modelID // Ensure model ID is set
			db.prices[modelID] = pricing
		}
	}
	
	return nil
}

// GetModelPricing returns the pricing for a specific model
func (db *PricingDatabase) GetModelPricing(modelID string) (*ModelPricing, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	// Try exact match first
	if pricing, exists := db.prices[modelID]; exists {
		return pricing, true
	}
	
	// Try without version suffix (e.g., "llama3" for "llama3:8b")
	if idx := strings.Index(modelID, ":"); idx > 0 {
		baseModel := modelID[:idx]
		if pricing, exists := db.prices[baseModel]; exists {
			return pricing, true
		}
	}
	
	// Try with lowercase
	lowerModel := strings.ToLower(modelID)
	if pricing, exists := db.prices[lowerModel]; exists {
		return pricing, true
	}
	
	return nil, false
}

// SetModelPricing sets or updates pricing for a model
func (db *PricingDatabase) SetModelPricing(pricing *ModelPricing) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.prices[pricing.ModelID] = pricing
}

// GetCurrency returns the current default currency
func (db *PricingDatabase) GetCurrency() Currency {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.currency
}

// SetCurrency sets the default currency
func (db *PricingDatabase) SetCurrency(currency Currency) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.currency = currency
}

// ConvertCurrency converts an amount from one currency to another
func (db *PricingDatabase) ConvertCurrency(amount float64, from, to Currency) float64 {
	if from == to {
		return amount
	}
	
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	// Convert to USD first if needed
	usdAmount := amount
	if from != USD {
		if rate, exists := db.rates[from]; exists && rate > 0 {
			usdAmount = amount / rate
		}
	}
	
	// Convert from USD to target currency
	if to != USD {
		if rate, exists := db.rates[to]; exists {
			return usdAmount * rate
		}
	}
	
	return usdAmount
}

// GetExchangeRate returns the exchange rate for a currency relative to USD
func (db *PricingDatabase) GetExchangeRate(currency Currency) float64 {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if rate, exists := db.rates[currency]; exists {
		return rate
	}
	return 1.0 // Default to USD rate
}

// ListModels returns all available model IDs with pricing
func (db *PricingDatabase) ListModels() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	models := make([]string, 0, len(db.prices))
	for modelID := range db.prices {
		models = append(models, modelID)
	}
	return models
}

// ExportConfig exports the current pricing configuration as JSON
func (db *PricingDatabase) ExportConfig() ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	config := struct {
		Currency string                   `json:"default_currency"`
		Rates    map[Currency]float64     `json:"exchange_rates"`
		Models   map[string]*ModelPricing `json:"model_pricing"`
	}{
		Currency: string(db.currency),
		Rates:    db.rates,
		Models:   db.prices,
	}
	
	return json.MarshalIndent(config, "", "  ")
}