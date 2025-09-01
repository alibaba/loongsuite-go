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

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()
	g := compose.NewGraph[[]*schema.Message, *schema.Message]()
	reactAgentKeyOfLambda, err := NewMockReActAgentLambda(ctx)
	if err != nil {
		panic(err)
	}
	err = g.AddLambdaNode("model", reactAgentKeyOfLambda)
	if err != nil {
		panic(err)
	}
	_ = g.AddEdge(compose.START, "model")
	_ = g.AddEdge("model", compose.END)
	graph, err := g.Compile(ctx)
	if err != nil {
		panic(err)
	}
	_, err = graph.Invoke(ctx, []*schema.Message{schema.UserMessage("hello")})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][3], "chat", "eino", "mock-chat")
		verifier.VerifyLLMCommonAttributes(stubs[0][6], "tool_node", "eino", trace.SpanKindClient)
		verifier.VerifyLLMCommonAttributes(stubs[0][7], "execute_tool", "eino", trace.SpanKindClient)
	}, 1)
}
