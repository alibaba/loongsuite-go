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

package verifier

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/alibaba/loongsuite-go-agent/pkg/testaccess"
	"github.com/mohae/deepcopy"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type node struct {
	root       bool
	childNodes []*node
	span       tracetest.SpanStub
}

func WaitAndAssertTraces(traceVerifiers func([]tracetest.SpanStubs), numTraces int) {
	traces := waitForTraces(numTraces)
	for i, trace := range traces {
		log.Printf("[test debugging] trace:%d\n", i)
		for _, span := range trace {
			log.Printf("[test debugging] " + span.Name)
			for _, attr := range span.Attributes {
				log.Printf("[test debugging] %v %v\n", attr.Key, attr.Value)
			}
		}
	}
	traceVerifiers(traces)
}

func WaitAndAssertMetrics(metricVerifiers map[string]func(metricdata.ResourceMetrics)) {
	mrs, err := waitForMetrics()
	log.Printf("[test debugging] %v\n", mrs)
	if err != nil {
		log.Fatalf("Failed to wait for metric: %v", err)
	}
	for k, v := range metricVerifiers {
		mrsCpy := DeepCopyMetric(mrs)
		d, err := filterMetricByName(mrsCpy, k)
		if err != nil {
			log.Fatalf("Failed to wait for metric: %v", err)
		}
		v(d)
	}
}

func waitForMetrics() (metricdata.ResourceMetrics, error) {
	var (
		mrs metricdata.ResourceMetrics
		err error
	)
	finish := false
	timeout := time.After(20 * time.Second)
	for !finish {
		select {
		case <-timeout:
			log.Fatalf("Timeout waiting for metrics!")
			finish = true
		default:
			time.Sleep(500 * time.Millisecond)
			mi, err := testaccess.GetTestMetrics()
			mrs = mi.(metricdata.ResourceMetrics)
			if err == nil {
				finish = true
				break
			}
		}
	}
	return mrs, err
}

func filterMetricByName(data metricdata.ResourceMetrics, name string) (metricdata.ResourceMetrics, error) {
	if len(data.ScopeMetrics) == 0 {
		return data, errors.New(fmt.Sprintf("No metrics named %s", name))
	}
	index := 0
	for _, s := range data.ScopeMetrics {
		scms := make([]metricdata.Metrics, 0)
		for j, sm := range s.Metrics {
			if sm.Name == name {
				scms = append(scms, s.Metrics[j])
			}
		}
		if len(scms) > 0 {
			data.ScopeMetrics[index].Metrics = scms
			index++
		}
	}
	return data, nil
}

func waitForTraces(numberOfTraces int) []tracetest.SpanStubs {
	defer testaccess.ResetTestSpans()
	finish := false
	var traces []tracetest.SpanStubs
	timeout := time.After(20 * time.Second)
	for !finish {
		select {
		case <-timeout:
			log.Printf("Timeout waiting for traces!")
			finish = true
		default:
			time.Sleep(500 * time.Millisecond)
			traces = groupAndSortTrace()
			if len(traces) >= numberOfTraces {
				finish = true
			}
		}
	}
	if len(traces) < numberOfTraces {
		log.Fatalf("Not enough traces, expected %d, got %d",
			numberOfTraces, len(traces))
	}
	return traces
}

func groupAndSortTrace() []tracetest.SpanStubs {
	spansi := testaccess.GetTestSpans()
	spans := spansi.(*tracetest.SpanStubs)
	traceMap := make(map[string][]tracetest.SpanStub)
	for _, span := range *spans {
		if span.SpanContext.HasTraceID() && span.SpanContext.TraceID().IsValid() {
			traceId := span.SpanContext.TraceID().String()
			spans, ok := traceMap[traceId]
			if !ok {
				spans = make([]tracetest.SpanStub, 0)
			}
			spans = append(spans, span)
			traceMap[traceId] = spans
		}
	}
	return sortTrace(traceMap)
}

func sortTrace(traceMap map[string][]tracetest.SpanStub) []tracetest.SpanStubs {
	traces := make([][]tracetest.SpanStub, 0)
	for _, trace := range traceMap {
		traces = append(traces, trace)
	}
	// ordered by start time
	sort.Slice(traces, func(i, j int) bool {
		return traces[i][0].StartTime.UnixNano() < traces[j][0].StartTime.UnixNano()
	})
	for i, _ := range traces {
		traces[i] = sortSingleTrace(traces[i])
	}
	stubs := make([]tracetest.SpanStubs, 0)
	for i, _ := range traces {
		stubs = append(stubs, traces[i])
	}
	return stubs
}

func sortSingleTrace(stubs []tracetest.SpanStub) []tracetest.SpanStub {
	// spans are ordered by their father-child relationship
	lookup := make(map[string]*node)
	for _, stub := range stubs {
		lookup[stub.SpanContext.SpanID().String()] = &node{
			root:       true,
			childNodes: make([]*node, 0),
			span:       stub,
		}
	}
	for _, stub := range stubs {
		n, ok := lookup[stub.SpanContext.SpanID().String()]
		if !ok {
			panic("no span id in stub " + stub.Name)
		}
		// find the parent node, then put it into parent node's childNodes
		if n.span.Parent.SpanID().IsValid() {
			parentSpanId := n.span.Parent.SpanID().String()
			parentNode, ok := lookup[parentSpanId]
			if ok {
				parentNode.childNodes = append(parentNode.childNodes, n)
				n.root = false
			}
		}
	}
	// find root
	rootNodes := make([]*node, 0)
	for _, stub := range stubs {
		n, ok := lookup[stub.SpanContext.SpanID().String()]
		if !ok {
			panic("no span id in stub " + stub.Name)
		}
		sort.Slice(n.childNodes, func(i, j int) bool {
			return n.childNodes[i].span.StartTime.UnixNano() < n.childNodes[j].span.StartTime.UnixNano()
		})
		if n.root {
			rootNodes = append(rootNodes, n)
		}
	}
	sort.Slice(rootNodes, func(i, j int) bool {
		return rootNodes[i].span.StartTime.UnixNano() < rootNodes[j].span.StartTime.UnixNano()
	})
	// walk the span tree
	t := make([]tracetest.SpanStub, 0)
	for _, rootNode := range rootNodes {
		traversePreOrder(rootNode, &t)
	}
	return t
}

func traversePreOrder(n *node, acc *[]tracetest.SpanStub) {
	*acc = append(*acc, n.span)
	for _, child := range n.childNodes {
		traversePreOrder(child, acc)
	}
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
