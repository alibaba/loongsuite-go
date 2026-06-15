// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package testaccess

import (
	"context"
	"os"

	"github.com/mohae/deepcopy"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const IS_IN_TEST = "IN_OTEL_TEST"

// In memory span exporter
var spanExporter = tracetest.NewInMemoryExporter()

var ManualReader = metric.NewManualReader()

func GetSpanExporter() trace.SpanExporter {
	return spanExporter
}

func GetTestSpans() interface{} {
	spans := spanExporter.GetSpans()
	return &spans
}

func ResetTestSpans() {
	spanExporter.Reset()
}

func GetTestMetrics() (interface{}, error) {
	var tmp, result metricdata.ResourceMetrics
	err := ManualReader.Collect(context.Background(), &tmp)
	if err != nil {
		return metricdata.ResourceMetrics{}, err
	}
	result = DeepCopyMetric(tmp)
	return result, nil
}

func DeepCopyMetric(mrs metricdata.ResourceMetrics) metricdata.ResourceMetrics {
	// do a deep copy in before each metric verifier executed
	mrsCpy := deepcopy.Copy(mrs).(metricdata.ResourceMetrics)
	// The deepcopy can not copy the attributes
	// so we just copy the data again to retain the attributes
	for i, s := range mrs.ScopeMetrics {
		for j, m := range s.Metrics {
			mrsCpy.ScopeMetrics[i].Metrics[j].Data = m.Data
		}
	}
	return mrsCpy
}

func IsInTest() bool {
	return os.Getenv(IS_IN_TEST) == "true"
}
