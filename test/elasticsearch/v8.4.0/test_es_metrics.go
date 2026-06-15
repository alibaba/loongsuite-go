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
	"strconv"

	"log"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var (
	client *elasticsearch.Client
	url    = "http://127.0.0.1:" + os.Getenv("OTEL_ES_PORT")
)

func main() {
	port, err := strconv.Atoi(os.Getenv("OTEL_ES_PORT"))
	if err != nil {
		panic(err)
	}

	client, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
		Password:  "123456",
		Username:  "elastic",
	})
	if err != nil {
		panic(err)
	}

	// creating an index
	_, err = client.Indices.Create("my_index")
	if err != nil {
		log.Printf("failed to create index %v\n", err)
	}

	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"http.client.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No http.client.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("http.client.request.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			verifier.VerifyHttpClientMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "PUT", "", "", "http", "1.1", port, 200)
		},
	})
}
