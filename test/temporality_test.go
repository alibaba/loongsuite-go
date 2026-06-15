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

package test

import (
	"encoding/json"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestTemporalityCumulative tests the default cumulative temporality
func TestTemporalityCumulative(t *testing.T) {
	UseApp("temporality")
	RunGoBuild(t, "go", "build", "test_temporality.go")

	env := []string{
		"OTEL_METRICS_EXPORTER=console",
		"OTEL_SERVICE_NAME=temporality-test-cumulative",
		"IN_OTEL_TEST=false", // Use real exporters instead of ManualReader
	}

	stdout, _ := RunApp(t, "test_temporality", env...)
	
	ExpectContains(t, stdout, "Temporality test completed successfully")
	
	verifyTemporality(t, stdout, "cumulative")
}

// TestTemporalityDelta tests the delta temporality preference
func TestTemporalityDelta(t *testing.T) {
	UseApp("temporality")
	RunGoBuild(t, "go", "build", "test_temporality.go")

	env := []string{
		"OTEL_METRICS_EXPORTER=console",
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta",
		"OTEL_SERVICE_NAME=temporality-test-delta",
		"IN_OTEL_TEST=false", // Use real exporters instead of ManualReader
	}

	stdout, _ := RunApp(t, "test_temporality", env...)
	
	ExpectContains(t, stdout, "Temporality test completed successfully")
	verifyTemporality(t, stdout, "delta")
}

// TestTemporalityLowMemory tests the lowmemory temporality preference
func TestTemporalityLowMemory(t *testing.T) {
	UseApp("temporality")
	RunGoBuild(t, "go", "build", "test_temporality.go")

	env := []string{
		"OTEL_METRICS_EXPORTER=console",
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=lowmemory",
		"OTEL_SERVICE_NAME=temporality-test-lowmemory",
		"IN_OTEL_TEST=false", // Use real exporters instead of ManualReader
	}

	stdout, _ := RunApp(t, "test_temporality", env...)
	
	ExpectContains(t, stdout, "Temporality test completed successfully")
	verifyTemporality(t, stdout, "lowmemory")
}

// TestTemporalityCaseInsensitive tests case-insensitive environment variable
func TestTemporalityCaseInsensitive(t *testing.T) {
	UseApp("temporality")
	RunGoBuild(t, "go", "build", "test_temporality.go")

	// Test with uppercase
	env := []string{
		"OTEL_METRICS_EXPORTER=console",
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA",
		"OTEL_SERVICE_NAME=temporality-test-uppercase",
		"IN_OTEL_TEST=false", // Use real exporters instead of ManualReader
	}

	stdout, _ := RunApp(t, "test_temporality", env...)
	ExpectContains(t, stdout, "Temporality test completed successfully")
	verifyTemporality(t, stdout, "delta")

	// Test with mixed case
	env = []string{
		"OTEL_METRICS_EXPORTER=console",
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=LowMemory",
		"OTEL_SERVICE_NAME=temporality-test-mixedcase",
		"IN_OTEL_TEST=false", // Use real exporters instead of ManualReader
	}

	stdout, _ = RunApp(t, "test_temporality", env...)
	ExpectContains(t, stdout, "Temporality test completed successfully")
	verifyTemporality(t, stdout, "lowmemory")
}

// verifyTemporality checks if the metrics have the expected temporality
func verifyTemporality(t *testing.T, output string, preference string) {
	foundCounter := false
	foundHistogram := false
	foundUpDownCounter := false
	
	lines := strings.Split(output, "\n")
	var jsonLine string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "{\"Resource\"") {
			jsonLine = line
			break
		}
	}
	
	if jsonLine == "" {
		t.Fatal("No JSON metrics output found")
	}
	
	var rm struct {
		ScopeMetrics []struct {
			Metrics []struct {
				Name string `json:"Name"`
				Data struct {
					Temporality string `json:"Temporality"`
				} `json:"Data"`
			} `json:"Metrics"`
		} `json:"ScopeMetrics"`
	}
	
	if err := json.Unmarshal([]byte(jsonLine), &rm); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, jsonLine)
	}
	
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			expectedTemporality := getExpectedTemporality(m.Name, preference)
			
			if m.Name == "test.counter" {
				foundCounter = true
				if m.Data.Temporality != expectedTemporality {
					t.Errorf("Counter temporality mismatch: got %s, want %s", 
						m.Data.Temporality, expectedTemporality)
				}
			} else if m.Name == "test.histogram" {
				foundHistogram = true
				if m.Data.Temporality != expectedTemporality {
					t.Errorf("Histogram temporality mismatch: got %s, want %s", 
						m.Data.Temporality, expectedTemporality)
				}
			} else if m.Name == "test.updowncounter" {
				foundUpDownCounter = true
				if m.Data.Temporality != expectedTemporality {
					t.Errorf("UpDownCounter temporality mismatch: got %s, want %s", 
						m.Data.Temporality, expectedTemporality)
				}
			}
		}
	}
	
	if !foundCounter {
		t.Error("test.counter metric not found in output")
	}
	if !foundHistogram {
		t.Error("test.histogram metric not found in output")
	}
	if !foundUpDownCounter {
		t.Error("test.updowncounter metric not found in output")
	}
}

// getExpectedTemporality returns the expected temporality string for a metric
func getExpectedTemporality(metricName string, preference string) string {
	switch preference {
	case "cumulative":
		return metricdata.CumulativeTemporality.String()
	case "delta":
		if metricName == "test.counter" || metricName == "test.histogram" {
			return metricdata.DeltaTemporality.String()
		}
		return metricdata.CumulativeTemporality.String()
	case "lowmemory":
		if metricName == "test.counter" || metricName == "test.histogram" {
			return metricdata.DeltaTemporality.String()
		}
		return metricdata.CumulativeTemporality.String()
	default:
		return metricdata.CumulativeTemporality.String()
	}
}
