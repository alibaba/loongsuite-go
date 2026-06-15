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

const http_server_request_duration = "http.server.request.duration"

const http_client_request_duration = "http.client.request.duration"

type HttpServerMetric struct {
	key                   attribute.Key
	serverRequestDuration metric.Float64Histogram
}

type HttpClientMetric struct {
	key                   attribute.Key
	clientRequestDuration metric.Float64Histogram
}

var mu sync.Mutex

var httpMetricsConv = map[attribute.Key]bool{
	semconv.HTTPRequestMethodKey:      true,
	semconv.URLSchemeKey:              true,
	semconv.ErrorTypeKey:              true,
	semconv.HTTPResponseStatusCodeKey: true,
	semconv.HTTPRouteKey:              true,
	semconv.NetworkProtocolNameKey:    true,
	semconv.NetworkProtocolVersionKey: true,
	semconv.ServerAddressKey:          true,
	semconv.ServerPortKey:             true,
}

var globalMeter metric.Meter

// InitHttpMetrics TODO: The init function may be executed after the HttpServerOperationListener() method
// so we need to make sure the otel_setup is executed before all the init() function
// related to issue https://github.com/alibaba/loongsuite-go-agent/issues/48
func InitHttpMetrics(m metric.Meter) {
	mu.Lock()
	defer mu.Unlock()
	globalMeter = m
}

func HttpServerMetrics(key string) *HttpServerMetric {
	mu.Lock()
	defer mu.Unlock()
	return &HttpServerMetric{key: attribute.Key(key)}
}

func HttpClientMetrics(key string) *HttpClientMetric {
	mu.Lock()
	defer mu.Unlock()
	return &HttpClientMetric{key: attribute.Key(key)}
}

// for test only
func newHttpServerMetric(key string, meter metric.Meter) (*HttpServerMetric, error) {
	m := &HttpServerMetric{
		key: attribute.Key(key),
	}
	d, err := newHttpServerRequestDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	m.serverRequestDuration = d
	return m, nil
}

func newHttpServerRequestDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(http_server_request_duration,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of HTTP server requests."))
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create http.server.request.duration histogram, %v", err))
	}
}

// for test only
func newHttpClientMetric(key string, meter metric.Meter) (*HttpClientMetric, error) {
	m := &HttpClientMetric{
		key: attribute.Key(key),
	}
	d, err := newHttpClientRequestDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	m.clientRequestDuration = d
	return m, nil
}

func newHttpClientRequestDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(http_client_request_duration,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of HTTP client requests."))
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create http.client.request.duration histogram, %v", err))
	}
}

type httpMetricContext struct {
	startTime       time.Time
	startAttributes []attribute.KeyValue
}

func (h *HttpServerMetric) OnBeforeStart(parentContext context.Context, startTime time.Time) context.Context {
	return parentContext
}

func (h *HttpServerMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTime time.Time) context.Context {
	return context.WithValue(ctx, h.key, httpMetricContext{
		startTime:       startTime,
		startAttributes: startAttributes,
	})
}

func (h *HttpServerMetric) OnAfterStart(context context.Context, endTime time.Time) {
	return
}

func (h *HttpServerMetric) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	mc := context.Value(h.key).(httpMetricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if h.serverRequestDuration == nil {
		var err error
		h.serverRequestDuration, err = newHttpServerRequestDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create serverRequestDuration, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)
	n, metricsAttrs := utils.Shadow(endAttributes, httpMetricsConv)
	if h.serverRequestDuration != nil {
		h.serverRequestDuration.Record(context, float64(endTime.Sub(startTime).Milliseconds()), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}
}

func (h HttpClientMetric) OnBeforeStart(parentContext context.Context, startTime time.Time) context.Context {
	return parentContext
}

func (h HttpClientMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTime time.Time) context.Context {
	return context.WithValue(ctx, h.key, httpMetricContext{
		startTime:       startTime,
		startAttributes: startAttributes,
	})
}

func (h HttpClientMetric) OnAfterStart(context context.Context, endTime time.Time) {
	return
}

func (h HttpClientMetric) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	mc := context.Value(h.key).(httpMetricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if h.clientRequestDuration == nil {
		var err error
		// second change to init the metric
		h.clientRequestDuration, err = newHttpClientRequestDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create clientRequestDuration, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)
	n, metricsAttrs := utils.Shadow(endAttributes, httpMetricsConv)
	if h.clientRequestDuration != nil {
		h.clientRequestDuration.Record(context, float64(endTime.Sub(startTime).Milliseconds()), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}
}
