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

package main

import (
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	api.InitDefault()
	flow.LoadRules([]*flow.Rule{
		{
			Resource:               "test",
			Threshold:              20,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
		{
			Resource:               "test_block",
			Threshold:              0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})
	e, _ := api.Entry(
		"test",
		api.WithResourceType(base.ResTypeWeb),
		api.WithTrafficType(base.Inbound),
	)
	if e != nil {
		e.Exit()
	}
	e, _ = api.Entry(
		"test_block",
		api.WithResourceType(base.ResTypeWeb),
		api.WithTrafficType(base.Inbound),
	)
	if e != nil {
		e.Exit()
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		// Find spans by resource name instead of assuming order
		var testSpan, testBlockSpan tracetest.SpanStub
		var foundTest, foundTestBlock bool
		
		// Iterate through all traces and spans
		for _, trace := range stubs {
			for _, span := range trace {
				resourceName := verifier.GetAttribute(span.Attributes, "sentinel.resource.name").AsString()
				if resourceName == "test" {
					testSpan = span
					foundTest = true
				} else if resourceName == "test_block" {
					testBlockSpan = span
					foundTestBlock = true
				}
			}
		}
		
		// Verify we found both spans
		if !foundTest {
			panic("Expected to find span with resource name 'test'")
		}
		if !foundTestBlock {
			panic("Expected to find span with resource name 'test_block'")
		}
		
		verifier.VerifySentinelAttributes(testSpan, "test", "Inbound", "", false)
		verifier.VerifySentinelAttributes(testBlockSpan, "test_block", "Inbound", "BlockTypeFlowControl", true)
	}, 2)
}
