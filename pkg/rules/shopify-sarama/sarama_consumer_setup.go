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

package sarama

import (
	"context"
	"strconv"
	_ "unsafe"

	"github.com/Shopify/sarama"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel"
)

//go:linkname consumePartitionOnEnter github.com/Shopify/sarama.consumePartitionOnEnter
func consumePartitionOnEnter(call api.CallContext, _ interface{}, topic string, partition int32, offset int64) {
}

//go:linkname consumePartitionOnExit github.com/Shopify/sarama.consumePartitionOnExit
func consumePartitionOnExit(call api.CallContext, pc sarama.PartitionConsumer, err error) {
	if !saramaEnabler.Enable() {
		return
	}
	if err != nil {
		return
	}

	wrapped := &tracingPartitionConsumer{
		PartitionConsumer: pc,
		messages:          make(chan *sarama.ConsumerMessage),
	}
	go wrapped.startFilter()

	call.SetReturnVal(0, wrapped)
}

func (tpc *tracingPartitionConsumer) startFilter() {
	msgs := tpc.PartitionConsumer.Messages()

	for msg := range msgs {
		offsetStr := strconv.FormatInt(msg.Offset, 10)
		partitionInt := int(msg.Partition)
		request := saramaConsumerRequest{
			msg:       msg,
			partition: &partitionInt,
			offset:    &offsetStr,
		}
		ctx := consumerInstrumenter.Start(context.Background(), request)

		otel.GetTextMapPropagator().Inject(ctx, saramaConsumerCarrier{msg: msg})

		tpc.messages <- msg

		consumerInstrumenter.End(ctx, request, nil, nil)
	}
	close(tpc.messages)
}
