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

	"github.com/IBM/sarama"
)

type saramaProducerRequest struct {
	msg        *sarama.ProducerMessage
	msgVersion sarama.KafkaVersion
}

type saramaProducerResponse struct {
	partition *int
	offset    *string
}

type saramaProducerMessageContext struct {
	traceCtx       context.Context
	msg            *sarama.ProducerMessage
	metadataBackup interface{}
}

type saramaConsumerRequest struct {
	msg       *sarama.ConsumerMessage
	partition *int
	offset    *string
}

type tracingAsyncProducer struct {
	sarama.AsyncProducer
	input         chan *sarama.ProducerMessage
	successes     chan *sarama.ProducerMessage
	errors        chan *sarama.ProducerError
	closeErr      chan error
	closeSig      chan struct{}
	closeAsyncSig chan struct{}
}

func (ap *tracingAsyncProducer) AsyncClose() {
	close(ap.input)
	close(ap.closeAsyncSig)
}

func (ap *tracingAsyncProducer) Close() error {
	close(ap.input)
	close(ap.closeSig)
	return <-ap.closeErr
}

func (ap *tracingAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	return ap.input
}

func (ap *tracingAsyncProducer) Successes() <-chan *sarama.ProducerMessage {
	return ap.successes
}

func (ap *tracingAsyncProducer) Errors() <-chan *sarama.ProducerError {
	return ap.errors
}

type tracingSyncProducer struct {
	sarama.SyncProducer
	saramaConfig *sarama.Config
}

type tracingPartitionConsumer struct {
	sarama.PartitionConsumer
	messages chan *sarama.ConsumerMessage
}

func (tpc *tracingPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return tpc.messages
}
