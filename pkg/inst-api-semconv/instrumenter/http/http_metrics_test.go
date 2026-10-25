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

package http

import (
	"context"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"testing"
	"time"
)

func TestHttpServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	server, err := newHttpServerMetric("test", meter)
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
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.server.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestHttpClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newHttpClientMetric("test", meter)
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
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

type metricEndListener interface {
	OnAfterEnd(ctx context.Context, endAttributes []attribute.KeyValue, endTime time.Time)
}

func newServerEndListener(t *testing.T, meter otelmetric.Meter) metricEndListener {
	server, err := newHttpServerMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func newClientEndListener(t *testing.T, meter otelmetric.Meter) metricEndListener {
	client, err := newHttpClientMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// An instrumenter can hand OnAfterEnd a context that never went through OnBeforeEnd, so the
// metric context has to be looked up defensively like DbClientMetric does.
func TestHttpMetricsOnAfterEndWithoutMetricContext(t *testing.T) {
	tests := []struct {
		name      string
		newClient bool
		ctx       context.Context
	}{
		{"server missing context", false, context.Background()},
		{"client missing context", true, context.Background()},
		{"server wrong context type", false, context.WithValue(context.Background(), attribute.Key("test"), "not-a-httpMetricContext")},
		{"client wrong context type", true, context.WithValue(context.Background(), attribute.Key("test"), "not-a-httpMetricContext")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := metric.NewManualReader()
			res := resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("my-service"),
				semconv.ServiceVersion("v0.1.0"),
			)
			mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
			var listener metricEndListener
			if tt.newClient {
				listener = newClientEndListener(t, mp.Meter("test-meter"))
			} else {
				listener = newServerEndListener(t, mp.Meter("test-meter"))
			}

			listener.OnAfterEnd(tt.ctx, []attribute.KeyValue{}, time.Now())

			rm := &metricdata.ResourceMetrics{}
			if err := reader.Collect(context.Background(), rm); err != nil {
				t.Fatal(err)
			}
			for _, sm := range rm.ScopeMetrics {
				for _, m := range sm.Metrics {
					t.Errorf("expected nothing to be recorded, got %s", m.Name)
				}
			}
		})
	}
}

func TestHttpMetricAttributesShadower(t *testing.T) {
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.KeyValue{
		Key:   semconv.HTTPRequestMethodKey,
		Value: attribute.StringValue("method"),
	}, attribute.KeyValue{
		Key:   "unknown",
		Value: attribute.Value{},
	}, attribute.KeyValue{
		Key:   semconv.NetworkProtocolNameKey,
		Value: attribute.StringValue("http"),
	}, attribute.KeyValue{
		Key:   semconv.ServerPortKey,
		Value: attribute.IntValue(8080),
	})
	n, attrs := utils.Shadow(attrs, httpMetricsConv)
	if n != 3 {
		panic("wrong shadow array")
	}
	if attrs[3].Key != "unknown" {
		panic("unknown should be the last attribute")
	}
}

func TestLazyHttpServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitHttpMetrics(m)
	server := HttpServerMetrics("net.http.server")
	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.server.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestLazyHttpClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitHttpMetrics(m)
	client := HttpClientMetrics("net.http.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestGlobalHttpServerMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitHttpMetrics(m)
	server := HttpServerMetrics("net.http.server")
	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.server.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestGlobalHttpClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitHttpMetrics(m)
	client := HttpClientMetrics("net.http.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "http.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestClientNilMeter(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	_ = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	_, err := newHttpClientMetric("test", nil)
	if err == nil {
		panic(err)
	}
}

func TestServerNilMeter(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	_ = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	_, err := newHttpServerMetric("test", nil)
	if err == nil {
		panic(err)
	}
}
