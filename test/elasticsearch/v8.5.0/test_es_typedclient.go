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
	"encoding/json"
	"log"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var (
	client *elasticsearch.TypedClient
	url    = "http://127.0.0.1:" + os.Getenv("OTEL_ES_PORT")
)

func main() {
	var err error
	client, err = elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{url},
		Password:  "123456",
		Username:  "elastic",
	})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	// creating an index
	_, err = client.Indices.Create("my_index").Do(ctx)
	if err != nil {
		log.Printf("failed to create index %v\n", err)
	}
	// indexing documents
	document := struct {
		Name string `json:"name"`
	}{
		"go-elasticsearch",
	}
	_, err = client.Index("my_index").
		Id("1").
		Request(document).
		Do(ctx)
	if err != nil {
		log.Printf("failed to index document %v\n", err)
	}
	// getting documents
	_, err = client.Get("my_index", "id").Do(ctx)
	if err != nil {
		log.Printf("failed to get documents %v\n", err)
	}
	// searching documents
	_, err = client.Search().Index("my_index").Request(&search.Request{Query: &types.Query{MatchAll: &types.MatchAllQuery{}}}).Do(ctx)
	if err != nil {
		log.Printf("failed to search documents %v\n", err)
	}
	// updating documents
	_, err = client.Update("my_index", "id").
		Request(&update.Request{
			Doc: json.RawMessage(`{ language: "Go" }`),
		}).Do(ctx)
	if err != nil {
		log.Printf("failed to update document %v\n", err)
	}
	// deleting documents
	_, err = client.Delete("my_index", "id").Do(ctx)
	if err != nil {
		log.Printf("failed to delete document %v\n", err)
	}
	// deleting an index
	_, err = client.Indices.Delete("my_index").Do(ctx)
	if err != nil {
		log.Printf("failed to delete index %v\n", err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "put", "elasticsearch", "127.0.0.1", "/my_index", "put", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "_doc", "elasticsearch", "127.0.0.1", "/my_index/_doc", "_doc", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "_doc", "elasticsearch", "127.0.0.1", "/my_index/_doc/id", "_doc", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "_search", "elasticsearch", "127.0.0.1", "/my_index/_search", "_search", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "_doc", "elasticsearch", "127.0.0.1", "/my_index/_doc/id", "_doc", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "delete", "elasticsearch", "127.0.0.1", "/my_index", "delete", "", nil)
	}, 1)
}
