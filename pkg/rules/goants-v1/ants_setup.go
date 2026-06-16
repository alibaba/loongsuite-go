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

package goants_v1

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/panjf2000/ants"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:linkname antsV1SubmitOnEnter github.com/panjf2000/ants.antsV1SubmitOnEnter
func antsV1SubmitOnEnter(call api.CallContext, p *ants.Pool, task func()) {
	if p == nil || task == nil {
		return
	}

	f := func() {
		opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindInternal)}
		_, span := otel.Tracer("github.com/panjf2000/ants").Start(context.Background(), "pool task", opts...)
		defer span.End()
		task()
	}
	call.SetParam(1, f)
}
