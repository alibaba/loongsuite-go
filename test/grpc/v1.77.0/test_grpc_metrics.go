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
	"context"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func main() {
	// starter server
	go setupGRPC()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	sendReq(context.Background())
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"rpc.server.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No rpc.server.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("rpc.server.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			verifier.VerifyRpcServerMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "Hello", "/HelloGrpc", "grpc", "127.0.0.1:9003")
		},
		"rpc.client.duration": func(rm metricdata.ResourceMetrics) {
			if len(rm.ScopeMetrics) <= 0 {
				panic("No rpc.client.duration metrics received!")
			}
			point := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("rpc.client.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			verifier.VerifyRpcClientMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "Hello", "/HelloGrpc", "grpc", "")
		},
	})

}
