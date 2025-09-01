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

package sentinel

import (
	"context"
	"log"
	"time"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/experimental"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OnEndHook struct {
	instrumenter *sentinelInstrumenter
}

var Hook = &OnEndHook{
	instrumenter: NewSentinelInstrumenter(),
}

func (h *OnEndHook) Order() uint32 {
	// execute at the end of slot chain
	return 10000
}

func (h *OnEndHook) OnEntryPassed(ctx *base.EntryContext) {
	// Do Nothing
}

func (h *OnEndHook) OnCompleted(ctx *base.EntryContext) {
	// collect sentinel attributes
	attrs := attribute.NewSet(attribute.KeyValue{
		Key:   experimental.ResourceName,
		Value: attribute.StringValue(ctx.Resource.Name()),
	}, attribute.KeyValue{
		Key:   experimental.EntryType,
		Value: attribute.StringValue(ctx.Resource.FlowType().String()),
	}, attribute.KeyValue{
		Key:   experimental.QPS,
		Value: attribute.Float64Value(ctx.StatNode.GetQPS(base.MetricEventPass)),
	}, attribute.KeyValue{
		Key:   experimental.BlockQps,
		Value: attribute.Float64Value(ctx.StatNode.GetQPS(base.MetricEventBlock)),
	}, attribute.KeyValue{
		Key:   experimental.AvgRT,
		Value: attribute.Float64Value(ctx.StatNode.AvgRT()),
	}, attribute.KeyValue{
		Key:   experimental.MinRT,
		Value: attribute.Float64Value(ctx.StatNode.MinRT()),
	}, attribute.KeyValue{
		Key:   experimental.RT,
		Value: attribute.Int64Value(int64(ctx.Rt())),
	}, attribute.KeyValue{
		Key:   experimental.IsBlocked,
		Value: attribute.BoolValue(ctx.IsBlocked()),
	})

	// calculate time
	StartTime := time.UnixMilli(int64(ctx.StartTime()))
	EndTime := StartTime.Add(time.Duration(ctx.Rt()) * time.Millisecond)

	h.instrumenter.StartAndEnd(context.Background(), ctx.Resource.Name(), StartTime, EndTime, attrs.ToSlice())
}

func (h *OnEndHook) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// collect sentinel attr
	attrs := attribute.NewSet(attribute.KeyValue{
		Key:   experimental.ResourceName,
		Value: attribute.StringValue(ctx.Resource.Name()),
	}, attribute.KeyValue{
		Key:   experimental.EntryType,
		Value: attribute.StringValue(ctx.Resource.FlowType().String()),
	}, attribute.KeyValue{
		Key:   experimental.QPS,
		Value: attribute.Float64Value(ctx.StatNode.GetQPS(base.MetricEventPass)),
	}, attribute.KeyValue{
		Key:   experimental.AvgRT,
		Value: attribute.Float64Value(ctx.StatNode.AvgRT()),
	}, attribute.KeyValue{
		Key:   experimental.MinRT,
		Value: attribute.Float64Value(ctx.StatNode.MinRT()),
	}, attribute.KeyValue{
		Key:   experimental.RT,
		Value: attribute.Int64Value(int64(ctx.Rt())),
	}, attribute.KeyValue{
		Key:   experimental.IsBlocked,
		Value: attribute.BoolValue(ctx.IsBlocked()),
	}, attribute.KeyValue{
		// Only supported when blocking
		Key:   experimental.BlockType,
		Value: attribute.StringValue(blockError.BlockType().String()),
	})
	// calculate time
	ctx.PutRt((uint64(time.Now().UnixMilli()) - ctx.StartTime()))
	StartTime := time.UnixMilli(int64(ctx.StartTime()))
	EndTime := StartTime.Add(time.Duration(ctx.Rt()) * time.Millisecond)

	// start and end span with time
	h.instrumenter.StartAndEnd(context.Background(), ctx.Resource.Name(), StartTime, EndTime, attrs.ToSlice())
}

//go:linkname onExitInitSentinel github.com/alibaba/sentinel-golang/api.onExitInitSentinel
func onExitInitSentinel(call api.CallContext, err error) {
	if !experimental.SentinelEnabler.Enable() {
		return
	}
	// add hook to slot chain
	sentinel.GlobalSlotChain().AddStatSlot(Hook)

	// register metric callback
	_, er := experimental.SentinelMeter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {

		// get all resource nodes
		Nodes := stat.ResourceNodeList()
		for _, node := range Nodes {

			// attribute set
			resName := node.ResourceName()
			attrSet := attribute.NewSet(attribute.KeyValue{
				Key:   experimental.ResourceName,
				Value: attribute.StringValue(resName),
			})

			// observe metric
			observer.ObserveFloat64(experimental.SentinelBlockQPS, node.GetQPS(base.MetricEventBlock), metric.WithAttributeSet(attrSet))
			observer.ObserveFloat64(experimental.SentinelPassQPS, node.GetQPS(base.MetricEventPass), metric.WithAttributeSet(attrSet))
		}
		return nil
	}, experimental.SentinelBlockQPS, experimental.SentinelPassQPS)
	if er != nil {
		log.Printf("[Sentinel] register metric callback failed: %v", er)
	}
}
