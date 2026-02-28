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
	"context"
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	_ kgo.HookProduceRecordBuffered   = new(KafkaTracer)
	_ kgo.HookProduceRecordUnbuffered = new(KafkaTracer)
	_ kgo.HookFetchRecordBuffered     = new(KafkaTracer)
	_ kgo.HookFetchRecordUnbuffered   = new(KafkaTracer)
)

type KafkaTracer struct{}

func (t *KafkaTracer) OnProduceRecordBuffered(r *kgo.Record) {
	if r.Context == nil {
		r.Context = context.Background()
	}
	request := franzKafkaProducerRequest{
		topic:   r.Topic,
		Record:  r,
		Context: r.Context,
	}
	ctx := franzKafkaProducerInstrumenter.Start(r.Context, request)
	r.Context = ctx
}

func (t *KafkaTracer) OnProduceRecordUnbuffered(r *kgo.Record, err error) {
	request := franzKafkaProducerRequest{
		topic:   r.Topic,
		Record:  r,
		Context: r.Context,
	}
	offsetStr := strconv.FormatInt(r.Offset, 10)
	response := franzKafkaProducerResponse{
		offset: &offsetStr,
	}
	franzKafkaProducerInstrumenter.End(r.Context, request, response, err)
}

func (t *KafkaTracer) OnFetchRecordBuffered(r *kgo.Record) {
	if r.Context == nil {
		r.Context = context.Background()
	}
	request := franzKafkaConsumerRequest{
		topic:   r.Topic,
		Record:  r,
		Context: r.Context,
	}
	newCtx := franzKafkaConsumerInstrumenter.Start(r.Context, request)
	r.Context = newCtx
}

func (t *KafkaTracer) OnFetchRecordUnbuffered(r *kgo.Record, _ bool) {
	request := franzKafkaConsumerRequest{
		topic:   r.Topic,
		Record:  r,
		Context: r.Context,
	}
	franzKafkaConsumerInstrumenter.End(r.Context, request, franzKafkaConsumerResponse{}, nil)
}
