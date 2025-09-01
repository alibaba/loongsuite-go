// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package experimental

import (
	"os"
	"testing"

	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/stretchr/testify/assert"
)

func TestInitSentinelExperimentalMetrics_GlobalMeterNil_NoMetricsInitialized(t *testing.T) {
	InitSentinelExperimentalMetrics(nil)
	assert.Nil(t, SentinelBlockQPS)
	assert.Nil(t, SentinelPassQPS)
}

func TestInitSentinelExperimentalMetrics_GlobalMeterNotNull_AllMetricsInitialized(t *testing.T) {
	mp := metric.NewMeterProvider()
	InitSentinelExperimentalMetrics(mp.Meter("a"))
	assert.NotNil(t, SentinelBlockQPS)
	assert.NotNil(t, SentinelPassQPS)
}

func TestSentinelEnablerDisable(t *testing.T) {
	se := sentinelEnabler{}
	os.Setenv("OTEL_INSTRUMENTATION_SENTINEL_EXPERIMENTAL_ENABLE", "false")
	if se.Enable() {
		panic("should not enable without OTEL_INSTRUMENTATION_SENTINEL_EXPERIMENTAL_ENABLE")
	}
}

func TestSentinelEnablerEnable(t *testing.T) {
	os.Setenv("OTEL_INSTRUMENTATION_SENTINEL_EXPERIMENTAL_ENABLE", "true")
	se := sentinelEnabler{}
	if !se.Enable() {
		panic("should enable with OTEL_INSTRUMENTATION_SENTINEL_EXPERIMENTAL_ENABLE")
	}
}
