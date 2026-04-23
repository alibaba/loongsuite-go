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
	cron "github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	c := cron.New(cron.WithSeconds())

	var once sync.Once
	done := make(chan struct{})
	_, err := c.AddFunc("@every 1s", func() {
		once.Do(func() { close(done) })
	})
	if err != nil {
		panic(err)
	}

	go c.Run()

	select {
	case <-done:
	case <-time.After(8 * time.Second):
		panic("cron task timeout")
	}

	stopCtx := c.Stop()
	<-stopCtx.Done()

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		var cronSpan, funcCronSpan tracetest.SpanStub
		for _, traceStubs := range stubs {
			for _, s := range traceStubs {
				if s.Name == "cron" {
					cronSpan = s
				}
				if s.Name == "func cron" {
					funcCronSpan = s
				}
			}
		}

		verifier.Assert(cronSpan.Name == "cron", "expected span name cron, got %s", cronSpan.Name)
		verifier.Assert(funcCronSpan.Name == "func cron", "expected span name func cron, got %s", funcCronSpan.Name)
		verifier.Assert(cronSpan.SpanKind == trace.SpanKindServer, "cron span kind should be server, got %d", cronSpan.SpanKind)
		verifier.Assert(funcCronSpan.SpanKind == trace.SpanKindServer, "func cron span kind should be server, got %d", funcCronSpan.SpanKind)
	}, 1)
}
