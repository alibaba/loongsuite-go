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

package asynq

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/arms_config"
	"github.com/alibaba/loongsuite-go-agent/pkg/go-sdk/otel"
	"github.com/alibaba/loongsuite-go-agent/pkg/go-sdk/otel/attribute"
	"github.com/alibaba/loongsuite-go-agent/pkg/go-sdk/otel/codes"
	"github.com/alibaba/loongsuite-go-agent/pkg/go-sdk/otel/trace"
	"github.com/hibiken/asynq"
)

func extractQueueFromOpts(opts []asynq.Option) string {
	for i := len(opts) - 1; i >= 0; i-- {
		if opts[i].Type() == asynq.QueueOpt {
			if v := opts[i].Value(); v != nil {
				if s, ok := v.(string); ok && s != "" {
					return s
				}
			}
		}
	}
	return "default"
}

//go:linkname asynqEnqueueContextOnEnter github.com/hibiken/asynq.asynqEnqueueContextOnEnter
func asynqEnqueueContextOnEnter(call api.CallContext, c *asynq.Client, ctx context.Context, task *asynq.Task, opts ...asynq.Option) {
	if !arms_config.ArmsEnable {
		return
	}
	if task == nil {
		return
	}

	optsSpan := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindProducer)}
	ctx, span := otel.ArmsTracer.Start(ctx, task.Type()+" enqueue", optsSpan...)

	queue := extractQueueFromOpts(opts)
	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "enqueue"),
		attribute.String("messaging.destination.name", queue),
		attribute.String("asynq.task.type", task.Type()),
		attribute.String("asynq.queue", queue),
	)

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["taskType"] = task.Type()
	data["queue"] = queue
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname asynqEnqueueContextOnExit github.com/hibiken/asynq.asynqEnqueueContextOnExit
func asynqEnqueueContextOnExit(call api.CallContext, info *asynq.TaskInfo, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "")
	if info != nil {
		span.SetAttributes(attribute.String("asynq.task.id", info.ID))
	}
}

//go:linkname asynqProcessTaskOnEnter github.com/hibiken/asynq.asynqProcessTaskOnEnter
func asynqProcessTaskOnEnter(call api.CallContext, mux *asynq.ServeMux, ctx context.Context, task *asynq.Task) {
	if !arms_config.ArmsEnable {
		return
	}
	if task == nil {
		return
	}

	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindConsumer)}
	ctx, span := otel.ArmsTracer.Start(ctx, task.Type()+" process", opts...)

	span.SetAttributes(
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "process"),
		attribute.String("asynq.task.type", task.Type()),
	)

	queue, _ := asynq.GetQueueName(ctx)
	if queue != "" {
		span.SetAttributes(
			attribute.String("messaging.destination.name", queue),
			attribute.String("asynq.queue", queue),
		)
	}

	retry, ok := asynq.GetRetryCount(ctx)
	if ok {
		span.SetAttributes(attribute.Int("asynq.retry_count", retry))
	}

	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["span"] = span
	data["taskType"] = task.Type()
	data["queue"] = queue
	call.SetData(data)
	call.SetParam(1, ctx)
}

//go:linkname asynqProcessTaskOnExit github.com/hibiken/asynq.asynqProcessTaskOnExit
func asynqProcessTaskOnExit(call api.CallContext, err error) {
	if call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	span, _ := data["span"].(trace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "")
}
