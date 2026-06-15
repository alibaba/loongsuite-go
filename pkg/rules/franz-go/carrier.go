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

package franz_go

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

// RecordCarrier injects and extracts traces from a kgo.Record.
// This type satisfies the otel/propagation.TextMapCarrier interface.
type RecordCarrier struct {
	record *kgo.Record
}

func NewRecordCarrier(record *kgo.Record) RecordCarrier {
	return RecordCarrier{record: record}
}

func (c RecordCarrier) Get(key string) string {
	for _, h := range c.record.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c RecordCarrier) Set(key, val string) {
	for i, h := range c.record.Headers {
		if h.Key == key {
			c.record.Headers[i].Value = []byte(val)
			return
		}
	}
	c.record.Headers = append(c.record.Headers, kgo.RecordHeader{
		Key:   key,
		Value: []byte(val),
	})
}

func (c RecordCarrier) Keys() []string {
	out := make([]string, len(c.record.Headers))
	for i, h := range c.record.Headers {
		out[i] = h.Key
	}
	return out
}
