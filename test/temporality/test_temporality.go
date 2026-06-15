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
	"time"

	"go.opentelemetry.io/otel"
)

func main() {
	meter := otel.Meter("test-meter")
	ctx := context.Background()

	counter, err := meter.Int64Counter("test.counter")
	if err != nil {
		panic(err)
	}

	histogram, err := meter.Float64Histogram("test.histogram")
	if err != nil {
		panic(err)
	}

	upDownCounter, err := meter.Int64UpDownCounter("test.updowncounter")
	if err != nil {
		panic(err)
	}

	counter.Add(ctx, 100)
	counter.Add(ctx, 50)

	histogram.Record(ctx, 1.5)
	histogram.Record(ctx, 2.5)

	upDownCounter.Add(ctx, 10)
	upDownCounter.Add(ctx, -5)

	// Wait for metrics to be collected and exported
	time.Sleep(3 * time.Second)

	fmt.Println("Temporality test completed successfully")
	fmt.Println("Check the exported metrics to verify temporality settings")
}
