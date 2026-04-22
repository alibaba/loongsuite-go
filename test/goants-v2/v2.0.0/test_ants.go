// Copyright (c) 2026 Alibaba Group Holding Ltd.
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
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	ants "github.com/panjf2000/ants/v2"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	p, err := ants.NewPool(1)
	if err != nil {
		panic(err)
	}
	defer p.Release()

	var wg sync.WaitGroup
	wg.Add(1)
	err = p.Submit(func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
	})
	if err != nil {
		panic(err)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		verifier.Assert(span.Name == "pool task", "expected span name pool task, got %s", span.Name)
		verifier.Assert(span.SpanKind == trace.SpanKindInternal, "expected internal span, got %d", span.SpanKind)
	}, 1)
}
