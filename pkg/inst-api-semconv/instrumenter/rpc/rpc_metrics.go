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

package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"log"
	"sync"
	"time"
)

const rpc_server_request_duration = "rpc.server.duration"

const rpc_client_request_duration = "rpc.client.duration"

type RpcServerMetric struct {
	key                   attribute.Key
	serverRequestDuration metric.Float64Histogram
}

type RpcClientMetric struct {
	key                   attribute.Key
	clientRequestDuration metric.Float64Histogram
}

var mu sync.Mutex

var rpcMetricsConv = map[attribute.Key]bool{
	semconv.RPCSystemKey:         true,
	semconv.RPCMethodKey:         true,
	semconv.RPCServiceKey:        true,
	semconv.ServerAddressKey:     true,
	semconv.RPCGRPCStatusCodeKey: true,
}

var globalMeter metric.Meter

// InitRpcMetrics so we need to make sure the otel_setup is executed before all the init() function
// related to issue rpcs://github.com/alibaba/loongsuite-go-agent/issues/48
func InitRpcMetrics(m metric.Meter) {
	mu.Lock()
	defer mu.Unlock()
	globalMeter = m
}

func RpcServerMetrics(key string) *RpcServerMetric {
	mu.Lock()
	defer mu.Unlock()
	return &RpcServerMetric{key: attribute.Key(key)}
}

func RpcClientMetrics(key string) *RpcClientMetric {
	mu.Lock()
	defer mu.Unlock()
	return &RpcClientMetric{key: attribute.Key(key)}
}

func newRpcServerRequestDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(rpc_server_request_duration,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of rpc server requests."))
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create rpc.server.request.duration histogram, %v", err))
	}
}

func newRpcClientRequestDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(rpc_client_request_duration,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of rpc client requests."))
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create rpc.client.request.duration histogram, %v", err))
	}
}

type rpcMetricContext struct {
	startTime       time.Time
	startAttributes []attribute.KeyValue
}

func (h *RpcServerMetric) OnBeforeStart(parentContext context.Context, startTime time.Time) context.Context {
	return parentContext
}

func (h *RpcServerMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTime time.Time) context.Context {
	return context.WithValue(ctx, h.key, rpcMetricContext{
		startTime:       startTime,
		startAttributes: startAttributes,
	})
}

func (h *RpcServerMetric) OnAfterStart(context context.Context, endTime time.Time) {
	return
}

func (h *RpcServerMetric) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	mc := context.Value(h.key).(rpcMetricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if h.serverRequestDuration == nil {
		var err error
		h.serverRequestDuration, err = newRpcServerRequestDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create serverRequestDuration, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)
	n, metricsAttrs := utils.Shadow(endAttributes, rpcMetricsConv)
	if h.serverRequestDuration != nil {
		h.serverRequestDuration.Record(context, float64(endTime.Sub(startTime).Milliseconds()), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}
}

func (h *RpcClientMetric) OnBeforeStart(parentContext context.Context, startTime time.Time) context.Context {
	return parentContext
}

func (h *RpcClientMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTime time.Time) context.Context {
	return context.WithValue(ctx, h.key, rpcMetricContext{
		startTime:       startTime,
		startAttributes: startAttributes,
	})
}

func (h *RpcClientMetric) OnAfterStart(context context.Context, endTime time.Time) {
	return
}

func (h *RpcClientMetric) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	if context.Value(h.key) == nil {
		return
	}
	mc := context.Value(h.key).(rpcMetricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if h.clientRequestDuration == nil {
		var err error
		// second change to init the metric
		h.clientRequestDuration, err = newRpcClientRequestDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create clientRequestDuration, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)

	n, metricsAttrs := utils.Shadow(endAttributes, rpcMetricsConv)
	if h.clientRequestDuration != nil {
		h.clientRequestDuration.Record(context, float64(endTime.Sub(startTime).Milliseconds()), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}
}

// for test only
func newRpcServerMetric(key string, meter metric.Meter) (*RpcServerMetric, error) {
	m := &RpcServerMetric{
		key: attribute.Key(key),
	}
	d, err := newRpcServerRequestDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	m.serverRequestDuration = d
	return m, nil
}

// for test only
func newRpcClientMetric(key string, meter metric.Meter) (*RpcClientMetric, error) {
	m := &RpcClientMetric{
		key: attribute.Key(key),
	}
	d, err := newRpcClientRequestDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	m.clientRequestDuration = d
	return m, nil
}
