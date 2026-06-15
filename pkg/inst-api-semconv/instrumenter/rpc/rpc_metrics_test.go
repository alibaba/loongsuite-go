// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      rpc://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"testing"
	"time"
)

func TestRpcServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	server, err := newRpcServerMetric("test", meter)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.server.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestRpcClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newRpcClientMetric("test", meter)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.client.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestRpcMetricAttributesShadower(t *testing.T) {
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.KeyValue{
		Key:   semconv.RPCMethodKey,
		Value: attribute.StringValue("method"),
	}, attribute.KeyValue{
		Key:   "unknown",
		Value: attribute.Value{},
	}, attribute.KeyValue{
		Key:   semconv.RPCServiceKey,
		Value: attribute.StringValue("rpc"),
	}, attribute.KeyValue{
		Key:   semconv.RPCSystemKey,
		Value: attribute.StringValue("abc"),
	})
	n, attrs := utils.Shadow(attrs, rpcMetricsConv)
	if n != 3 {
		panic("wrong shadow array")
	}
	if attrs[3].Key != "unknown" {
		panic("unknown should be the last attribute")
	}
}

func TestLazyRpcServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitRpcMetrics(m)
	server := RpcServerMetrics("net.rpc.server")
	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.server.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestLazyRpcClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitRpcMetrics(m)
	client := RpcClientMetrics("net.rpc.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.client.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestGlobalRpcServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitRpcMetrics(m)
	server := RpcServerMetrics("net.rpc.server")
	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.server.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestGlobalRpcClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitRpcMetrics(m)
	client := RpcClientMetrics("net.rpc.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "rpc.client.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestNilClientMeter(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	_ = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	_, err := newRpcClientMetric("test", nil)
	if err == nil {
		panic(err)
	}
}

func TestNilServerMeter(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	_ = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	_, err := newRpcServerMetric("test", nil)
	if err == nil {
		panic(err)
	}
}
