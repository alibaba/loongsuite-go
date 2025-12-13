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

package mqtt

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

func Enabled() bool { return mqttEnabler.Enable() }

func ProducerInstrumenter() interface {
	Start(context.Context, PublishRequest, ...trace.SpanStartOption) context.Context
	End(context.Context, PublishRequest, PublishResponse, error, ...trace.SpanEndOption)
} {
	return publishInst
}

func ConsumerInstrumenter() interface {
	Start(context.Context, DeliverRequest, ...trace.SpanStartOption) context.Context
	End(context.Context, DeliverRequest, DeliverResponse, error, ...trace.SpanEndOption)
} {
	return deliverInst
}

func StartProducer(ctx context.Context, req PublishRequest) context.Context {
	return StartPublish(ctx, req)
}
func EndProducer(ctx context.Context, req PublishRequest, res PublishResponse, err error) {
	EndPublish(ctx, req, res, err)
}

func StartConsumer(ctx context.Context, req DeliverRequest) context.Context {
	return StartDeliver(ctx, req)
}
func EndConsumer(ctx context.Context, req DeliverRequest, res DeliverResponse, err error) {
	EndDeliver(ctx, req, res, err)
}
