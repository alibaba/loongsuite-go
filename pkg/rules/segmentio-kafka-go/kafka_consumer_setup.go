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

package kafka

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/segmentio/kafka-go"
	"time"
	_ "unsafe"
)

//go:linkname consumerReadMessageOnEnter github.com/segmentio/kafka-go.consumerReadMessageOnEnter
func consumerReadMessageOnEnter(call api.CallContext, _ interface{}, ctx context.Context) {
	if !kafkaEnabler.Enable() {
		return
	}

	instrumentationData := map[string]interface{}{
		"parentContext":  ctx,
		"startTimestamp": time.Now(),
	}
	call.SetData(instrumentationData)
}

//go:linkname consumerReadMessageOnExit github.com/segmentio/kafka-go.consumerReadMessageOnExit
func consumerReadMessageOnExit(call api.CallContext, message kafka.Message, err error) {
	if !kafkaEnabler.Enable() {
		return
	}

	instrumentationData, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	parentContext := instrumentationData["parentContext"].(context.Context)
	startTimestamp := instrumentationData["startTimestamp"].(time.Time)
	endTimestamp := time.Now()

	consumerRequest := kafkaConsumerReq{msg: message}
	consumerInstrumenter.StartAndEnd(
		parentContext,
		consumerRequest,
		nil,
		err,
		startTimestamp,
		endTimestamp,
	)
}
