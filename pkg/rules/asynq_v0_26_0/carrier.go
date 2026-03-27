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

package asynq_v0_26_0

import (
	"github.com/hibiken/asynq"
)

// TaskCarrier injects and extracts traces from an asynq.Task's headers.
type TaskCarrier struct {
	task *asynq.Task
}

func NewTaskCarrier(task *asynq.Task) TaskCarrier {
	return TaskCarrier{task: task}
}

func (c TaskCarrier) Get(key string) string {
	if c.task.Headers() == nil {
		return ""
	}
	return c.task.Headers()[key]
}

func (c TaskCarrier) Set(key, val string) {
	header := c.task.Headers()
	if header == nil {
		header = make(map[string]string)
	}
	header[key] = val
	c.task.SetHeaders(header)
}

func (c TaskCarrier) Keys() []string {
	headers := c.task.Headers()
	if len(headers) == 0 {
		return nil
	}
	out := make([]string, 0, len(headers))
	for k := range headers {
		out = append(out, k)
	}
	return out
}
