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
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// GetRelevantDocuments
//
//go:linkname getRelevantDocumentsOnEnter github.com/tmc/langchaingo/vectorstores.getRelevantDocumentsOnEnter
func getRelevantDocumentsOnEnter(call api.CallContext,
	r vectorstores.Retriever,
	ctx context.Context,
	query string,
) {
	request := langChainRequest{
		operationName: MRelevantDoc,
		system:        "langchain",
		spanKind:      ai.GenAISpanKindRetriever,
	}
	langCtx := langChainCommonInstrument.Start(ctx, request)
	data := make(map[string]interface{})
	data["ctx"] = langCtx
	call.SetData(data)
}

//go:linkname getRelevantDocumentsOnExit github.com/tmc/langchaingo/vectorstores.getRelevantDocumentsOnExit
func getRelevantDocumentsOnExit(
	call api.CallContext,
	s []schema.Document, err error,
) {
	data := call.GetData().(map[string]interface{})
	request := langChainRequest{
		operationName: MRelevantDoc,
		system:        "langchain",
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	langChainCommonInstrument.End(ctx, request, nil, err)
}
