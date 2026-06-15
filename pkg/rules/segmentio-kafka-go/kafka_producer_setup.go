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
	_ "unsafe"
)

//go:linkname producerWriteMessagesOnEnter github.com/segmentio/kafka-go.producerWriteMessagesOnEnter
func producerWriteMessagesOnEnter(call api.CallContext, writer *kafka.Writer, ctx context.Context, messages ...kafka.Message) {
	if !kafkaEnabler.Enable() {
		return
	}

	// Create pointers to messages for instrumentation
	messagePointers := make([]*kafka.Message, len(messages))
	for i := range messages {
		// Create a copy to avoid issues with loop variable
		messageCopy := messages[i]
		messagePointers[i] = &messageCopy
	}

	// Prepare request data for instrumentation
	producerRequest := kafkaProducerReq{
		topic: writer.Topic,
		addr:  writer.Addr,
		async: writer.Async,
		msgs:  messagePointers,
	}

	// Start instrumentation and get instrumented context
	instrumentedContext := producerInstrumenter.Start(ctx, producerRequest)

	// Store data for later use in exit hook
	instrumentationData := map[string]interface{}{
		"instrumentedContext": instrumentedContext,
		"producerRequest":     producerRequest,
	}
	call.SetData(instrumentationData)

	// Prepare message copies to maintain idempotency
	messageCopies := make([]kafka.Message, len(messagePointers))
	for i, msg := range messagePointers {
		messageCopies[i] = *msg
	}

	// Update the call parameters with copied messages
	call.SetParam(2, messageCopies)
}

//go:linkname producerWriteMessagesOnExit github.com/segmentio/kafka-go.producerWriteMessagesOnExit
func producerWriteMessagesOnExit(call api.CallContext, err error) {
	if !kafkaEnabler.Enable() {
		return
	}

	// Retrieve stored instrumentation data
	instrumentationData := call.GetData().(map[string]interface{})
	instrumentedContext := instrumentationData["instrumentedContext"].(context.Context)
	producerRequest := instrumentationData["producerRequest"].(kafkaProducerReq)

	// End instrumentation with results
	producerInstrumenter.End(instrumentedContext, producerRequest, nil, err)
}
