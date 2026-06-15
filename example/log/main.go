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

package main

import (
	"net/http"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func main() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		logger := zap.NewExample()
		logger.Debug("this is debug message")
		logger.Info("this is info message")
		logger.Info("this is info message with fields",
			zap.Int("age", 37),
			zap.String("agender", "man"),
		)
		logger.Warn("this is warn message")
		logger.Error("this is error message")
	})

	http.HandleFunc("/logwithtrace", func(w http.ResponseWriter, r *http.Request) {
		logger := zap.NewExample()
		// GetTraceAndSpanId will be added while using otel, users must use otel to build the module
		traceId, spanId := trace.GetTraceAndSpanId()
		logger.Info("this is info message with fields",
			zap.String("traceId", traceId),
			zap.String("spanId", spanId),
		)
	})
	http.ListenAndServe(":9999", nil)
}
