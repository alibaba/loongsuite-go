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

package langchain

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/tmc/langchaingo/embeddings"
)

// EmbedQuery
//
//go:linkname singleEmbedOnEnter github.com/tmc/langchaingo/embeddings.singleEmbedOnEnter
func singleEmbedOnEnter(call api.CallContext,
	e *embeddings.EmbedderImpl,
	ctx context.Context,
	text string,
) {
	request := langChainRequest{
		operationName: MEmbedSingle,
		system:        "langchain",
		spanKind:      ai.GenAISpanKindEmbedding,
	}
	langCtx := langChainCommonInstrument.Start(ctx, request)
	data := make(map[string]interface{})
	data["ctx"] = langCtx
	call.SetData(data)
}

//go:linkname singleEmbedOnExit github.com/tmc/langchaingo/embeddings.singleEmbedOnExit
func singleEmbedOnExit(
	call api.CallContext,
	emb []float32,
	err error,
) {
	request := langChainRequest{
		operationName: MEmbedSingle,
		system:        "langchain",
	}
	data := call.GetData().(map[string]interface{})
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	langChainCommonInstrument.End(ctx, request, nil, err)
}

// BatchedEmbed
//
//go:linkname batchedEmbedOnEnter github.com/tmc/langchaingo/embeddings.batchedEmbedOnEnter
func batchedEmbedOnEnter(call api.CallContext,
	ctx context.Context,
	embedder embeddings.EmbedderClient,
	texts []string,
	batchSize int,
) {
	request := langChainRequest{
		operationName: MEmbedBatch,
		system:        "langchain",
		spanKind:      ai.GenAISpanKindEmbedding,
		input: map[string]interface{}{
			"batchSize": batchSize,
		},
	}
	langCtx := langChainCommonInstrument.Start(ctx, request)
	data := make(map[string]interface{})
	data["ctx"] = langCtx
	call.SetData(data)
}

//go:linkname batchedEmbedOnExit github.com/tmc/langchaingo/embeddings.batchedEmbedOnExit
func batchedEmbedOnExit(
	call api.CallContext,
	emb [][]float32,
	err error,
) {
	request := langChainRequest{
		operationName: MEmbedBatch,
		system:        "langchain",
	}
	data := call.GetData().(map[string]interface{})
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	langChainCommonInstrument.End(ctx, request, nil, err)
}
