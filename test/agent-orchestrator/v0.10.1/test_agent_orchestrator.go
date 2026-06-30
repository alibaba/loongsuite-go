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
	"context"
	"log"
	"strings"

	"github.com/aoagents/agent-orchestrator/backend/internal/session_manager"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	mgr := sessionmanager.New()

	rec, err := mgr.Spawn(context.Background(), sessionmanager.SpawnConfig{
		ProjectID: "proj-42",
		IssueID:   "issue-7",
		Kind:      sessionmanager.KindOrchestrator,
		Harness:   "claude-code",
		Branch:    "feat/cool-thing",
		Prompt:    "implement the feature",
	})
	if err != nil {
		log.Fatalf("spawn failed: %v", err)
	}

	if err := mgr.Send(context.Background(), rec.ID, "ping"); err != nil {
		log.Fatalf("send failed: %v", err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		var spawnSpan, sendSpan tracetest.SpanStub
		for _, traceStubs := range stubs {
			for _, s := range traceStubs {
				switch s.Name {
				case "agent-orchestrator spawn":
					spawnSpan = s
				case "agent-orchestrator send":
					sendSpan = s
				}
			}
		}

		verifier.Assert(spawnSpan.Name == "agent-orchestrator spawn",
			"expected spawn span, got %s", spawnSpan.Name)
		verifier.Assert(spawnSpan.SpanKind == trace.SpanKindInternal,
			"spawn span kind should be internal, got %d", spawnSpan.SpanKind)
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.system") == "agent_orchestrator",
			"spawn gen_ai.system mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.operation.name") == "spawn_agent",
			"spawn gen_ai.operation.name mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.span.kind") == "workflow",
			"spawn gen_ai.span.kind mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.other_input.agent_harness") == "claude-code",
			"spawn agent_harness mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.other_input.session_kind") == "orchestrator",
			"spawn session_kind mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.other_input.project_id") == "proj-42",
			"spawn project_id mismatch")
		verifier.Assert(attrVal(spawnSpan.Attributes, "gen_ai.other_input.branch") == "feat/cool-thing",
			"spawn branch mismatch")
		verifier.Assert(strings.HasPrefix(attrVal(spawnSpan.Attributes, "gen_ai.other_output.session_id"), "sess-claude-code-"),
			"spawn session_id mismatch: %s", attrVal(spawnSpan.Attributes, "gen_ai.other_output.session_id"))

		verifier.Assert(sendSpan.Name == "agent-orchestrator send",
			"expected send span, got %s", sendSpan.Name)
		verifier.Assert(sendSpan.SpanKind == trace.SpanKindClient,
			"send span kind should be client, got %d", sendSpan.SpanKind)
		verifier.Assert(attrVal(sendSpan.Attributes, "gen_ai.system") == "agent_orchestrator",
			"send gen_ai.system mismatch")
		verifier.Assert(attrVal(sendSpan.Attributes, "gen_ai.operation.name") == "send_message",
			"send gen_ai.operation.name mismatch")
		verifier.Assert(attrVal(sendSpan.Attributes, "gen_ai.span.kind") == "client",
			"send gen_ai.span.kind mismatch")
		verifier.Assert(strings.HasPrefix(attrVal(sendSpan.Attributes, "gen_ai.other_input.session_id"), "sess-claude-code-"),
			"send session_id mismatch")
	}, 1)
}

func attrVal(attrs []attribute.KeyValue, key string) string {
	for _, a := range attrs {
		if string(a.Key) == key {
			return a.Value.AsString()
		}
	}
	return ""
}
