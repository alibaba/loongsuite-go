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
	"os"
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const taskType = "email:welcome"
const queueName = "default"

func main() {
	redisAddr := "localhost:" + os.Getenv("REDIS_PORT")
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	srv := asynq.NewServer(redisOpt, asynq.Config{Concurrency: 1})
	mux := asynq.NewServeMux()
	mux.HandleFunc(taskType, func(ctx context.Context, t *asynq.Task) error {
		defer wg.Done()
		return nil
	})

	go func() {
		_ = srv.Run(mux)
	}()

	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()
	task := asynq.NewTask(taskType, []byte(`{"user_id":42}`))
	_, err := client.EnqueueContext(ctx, task)
	if err != nil {
		panic(err)
	}

	wg.Wait()
	time.Sleep(2 * time.Second)

	verifier.WaitAndAssertTracesWithNum(func(stubs []tracetest.SpanStubs) {
		var enqueueSpan, processSpan tracetest.SpanStub
		for _, traceStubs := range stubs {
			for _, s := range traceStubs {
				if s.SpanKind == trace.SpanKindProducer {
					enqueueSpan = s
				} else if s.SpanKind == trace.SpanKindConsumer {
					processSpan = s
				}
			}
		}

		verifier.Assert(enqueueSpan.Name != "", "expected to find enqueue (producer) span")
		verifier.Assert(processSpan.Name != "", "expected to find process (consumer) span")
		verifier.Assert(enqueueSpan.SpanKind == trace.SpanKindProducer, "enqueue span should be producer, got %d", enqueueSpan.SpanKind)
		verifier.Assert(processSpan.SpanKind == trace.SpanKindConsumer, "process span should be consumer, got %d", processSpan.SpanKind)

		verifier.Assert(enqueueSpan.Name == taskType+" enqueue", "expected enqueue span name %s enqueue, got %s", taskType, enqueueSpan.Name)
		verifier.Assert(processSpan.Name == taskType+" process", "expected process span name %s process, got %s", taskType, processSpan.Name)

		verifier.Assert(enqueueSpan.SpanContext.TraceID().String() == processSpan.SpanContext.TraceID().String(), "v0.26.0: enqueue and process spans should be in same trace (header propagation)")

		sys := verifier.GetAttribute(enqueueSpan.Attributes, "messaging.system").AsString()
		verifier.Assert(sys == "asynq", "expected messaging.system asynq, got %s", sys)
		sys = verifier.GetAttribute(processSpan.Attributes, "messaging.system").AsString()
		verifier.Assert(sys == "asynq", "expected messaging.system asynq, got %s", sys)

		op := verifier.GetAttribute(enqueueSpan.Attributes, "messaging.operation").AsString()
		verifier.Assert(op == "enqueue", "expected messaging.operation enqueue, got %s", op)
		op = verifier.GetAttribute(processSpan.Attributes, "messaging.operation").AsString()
		verifier.Assert(op == "process", "expected messaging.operation process, got %s", op)

		tt := verifier.GetAttribute(enqueueSpan.Attributes, "asynq.task.type").AsString()
		verifier.Assert(tt == taskType, "expected asynq.task.type %s, got %s", taskType, tt)
		tt = verifier.GetAttribute(processSpan.Attributes, "asynq.task.type").AsString()
		verifier.Assert(tt == taskType, "expected asynq.task.type %s, got %s", taskType, tt)

		queue := verifier.GetAttribute(enqueueSpan.Attributes, "asynq.queue").AsString()
		verifier.Assert(queue == queueName, "expected asynq.queue %s, got %s", queueName, queue)
		queue = verifier.GetAttribute(processSpan.Attributes, "asynq.queue").AsString()
		verifier.Assert(queue == queueName, "expected asynq.queue %s, got %s", queueName, queue)
	}, 1)
}
