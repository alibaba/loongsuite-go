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

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()

	loader := &MockLoader{}
	embedder := &MockEmbedder{}
	mockIndexer := NewMockIndexer()
	mockRetriever := NewMockRetriever(mockIndexer.storage)

	graph := compose.NewGraph[string, []*schema.Document]()

	err := graph.AddLoaderNode("loader", loader)
	if err != nil {
		panic(err)
	}
	err = graph.AddEmbeddingNode("embedder", embedder)
	if err != nil {
		panic(err)
	}
	err = graph.AddIndexerNode("indexer", mockIndexer)
	if err != nil {
		panic(err)
	}
	err = graph.AddRetrieverNode("retriever", mockRetriever)
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("doc_to_text", compose.InvokableLambda(
		func(ctx context.Context, docs []*schema.Document) ([]string, error) {
			texts := make([]string, len(docs))
			for i, doc := range docs {
				texts[i] = doc.Content
			}
			return texts, nil
		}))
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("query_input", compose.InvokableLambda(
		func(ctx context.Context, query string) (document.Source, error) {
			return document.Source{}, nil
		}))
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("prepare_for_index", compose.InvokableLambda(
		func(ctx context.Context, embeddings [][]float64) ([]*schema.Document, error) {
			return []*schema.Document{
				{ID: "doc1", Content: "这是第一个文档的内容"},
				{ID: "doc2", Content: "这是第二个文档的内容"},
				{ID: "doc3", Content: "这是第三个文档的内容"},
			}, nil
		}))
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("create_query", compose.InvokableLambda(
		func(ctx context.Context, ids []string) (string, error) {
			return "文档", nil
		}))
	if err != nil {
		panic(err)
	}

	_ = graph.AddEdge(compose.START, "query_input")
	_ = graph.AddEdge("query_input", "loader")
	_ = graph.AddEdge("loader", "doc_to_text")
	_ = graph.AddEdge("doc_to_text", "embedder")
	_ = graph.AddEdge("embedder", "prepare_for_index")
	_ = graph.AddEdge("prepare_for_index", "indexer")
	_ = graph.AddEdge("indexer", "create_query")
	_ = graph.AddEdge("create_query", "retriever")
	_ = graph.AddEdge("retriever", compose.END)

	runnable, err := graph.Compile(ctx, compose.WithMaxRunSteps(100))
	if err != nil {
		panic(err)
	}

	_, err = runnable.Invoke(ctx, "测试查询")
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMCommonAttributes(stubs[0][0], "graph", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][1], "lambda", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][2], "loader", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][3], "lambda", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][4], "embeddings", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][5], "lambda", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][6], "indexer", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][7], "lambda", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][8], "retriever", "eino", trace.SpanKindClient)
	}, 1)
}
