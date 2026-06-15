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

package instrumenter

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"testing"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNoopSpanSuppressor(t *testing.T) {
	ns := &NoneStrategy{}
	n := ns.create([]attribute.Key{})
	ctx := context.Background()
	n.StoreInContext(ctx, trace.SpanKindClient, noop.Span{})
	if n.ShouldSuppress(ctx, trace.SpanKindClient) != false {
		t.Errorf("should not suppress span")
	}
}

func TestSpanKeySuppressor(t *testing.T) {
	s := SpanKeySuppressor{
		spanKeys: []attribute.Key{
			utils.HTTP_CLIENT_KEY,
		},
	}
	builder := Builder[testRequest, testResponse]{}
	builder.Init().SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:      utils.FAST_HTTP_CLIENT_SCOPE_NAME,
			Version:   "test",
			SchemaURL: "test",
		})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	traceProvider := sdktrace.NewTracerProvider()
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(traceProvider)
	defer otel.SetTracerProvider(originalTP)
	newCtx := instrumenter.Start(ctx, testRequest{})
	span := trace.SpanFromContext(newCtx)
	newCtx = s.StoreInContext(newCtx, trace.SpanKindClient, span)
	if !s.ShouldSuppress(newCtx, trace.SpanKindClient) {
		t.Errorf("should suppress span")
	}
}

func TestSpanKeySuppressorNotMatch(t *testing.T) {
	s := SpanKeySuppressor{
		spanKeys: []attribute.Key{
			utils.RPC_CLIENT_KEY,
		},
	}
	builder := Builder[testRequest, testResponse]{}
	builder.Init().SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:      utils.FAST_HTTP_CLIENT_SCOPE_NAME,
			Version:   "test",
			SchemaURL: "test",
		})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	traceProvider := sdktrace.NewTracerProvider()
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(traceProvider)
	defer otel.SetTracerProvider(originalTP)
	newCtx := instrumenter.Start(ctx, testRequest{})
	span := trace.SpanFromContext(newCtx)
	newCtx = s.StoreInContext(newCtx, trace.SpanKindClient, span)
	if s.ShouldSuppress(newCtx, trace.SpanKindClient) {
		t.Errorf("should not suppress span with different span key")
	}
}

func TestSpanKindSuppressor(t *testing.T) {
	sks := &SpanKindStrategy{}
	s := sks.create([]attribute.Key{})
	builder := Builder[testRequest, testResponse]{}
	builder.Init().SetSpanNameExtractor(testNameExtractor{}).
		SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:      utils.FAST_HTTP_CLIENT_SCOPE_NAME,
			Version:   "test",
			SchemaURL: "test",
		})
	instrumenter := builder.BuildInstrumenter()
	ctx := context.Background()
	traceProvider := sdktrace.NewTracerProvider()
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(traceProvider)
	defer otel.SetTracerProvider(originalTP)
	newCtx := instrumenter.Start(ctx, testRequest{})
	span := trace.SpanFromContext(newCtx)
	newCtx = s.StoreInContext(newCtx, trace.SpanKindClient, span)
	if !s.ShouldSuppress(newCtx, trace.SpanKindClient) {
		t.Errorf("should not suppress span with different span key")
	}
}

func TestGetScopeKey_HTTP(t *testing.T) {
	tests := []struct {
		name      string
		scopeName string
		spanKind  trace.SpanKind
		expected  attribute.Key
	}{
		{
			name:      "nethttp client",
			scopeName: "loongsuite.instrumentation.nethttp",
			spanKind:  trace.SpanKindClient,
			expected:  utils.HTTP_CLIENT_KEY,
		},
		{
			name:      "nethttp server",
			scopeName: "loongsuite.instrumentation.nethttp",
			spanKind:  trace.SpanKindServer,
			expected:  utils.HTTP_SERVER_KEY,
		},
		{
			name:      "fasthttp client",
			scopeName: "loongsuite.instrumentation.fasthttp",
			spanKind:  trace.SpanKindClient,
			expected:  utils.HTTP_CLIENT_KEY,
		},
		{
			name:      "fasthttp server",
			scopeName: "loongsuite.instrumentation.fasthttp",
			spanKind:  trace.SpanKindServer,
			expected:  utils.HTTP_SERVER_KEY,
		},
		{
			name:      "hertz client",
			scopeName: "loongsuite.instrumentation.hertz",
			spanKind:  trace.SpanKindClient,
			expected:  utils.HTTP_CLIENT_KEY,
		},
		{
			name:      "hertz server",
			scopeName: "loongsuite.instrumentation.hertz",
			spanKind:  trace.SpanKindServer,
			expected:  utils.HTTP_SERVER_KEY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getScopeKey(tt.scopeName, tt.spanKind)
			if result != tt.expected {
				t.Errorf("getScopeKey(%s, %v) = %v, want %v", tt.scopeName, tt.spanKind, result, tt.expected)
			}
		})
	}
}

func TestGetScopeKey_RPC(t *testing.T) {
	tests := []struct {
		name      string
		scopeName string
		spanKind  trace.SpanKind
		expected  attribute.Key
	}{
		{
			name:      "grpc client",
			scopeName: "loongsuite.instrumentation.grpc",
			spanKind:  trace.SpanKindClient,
			expected:  utils.RPC_CLIENT_KEY,
		},
		{
			name:      "grpc server",
			scopeName: "loongsuite.instrumentation.grpc",
			spanKind:  trace.SpanKindServer,
			expected:  utils.RPC_SERVER_KEY,
		},
		{
			name:      "trpc client",
			scopeName: "loongsuite.instrumentation.trpc",
			spanKind:  trace.SpanKindClient,
			expected:  utils.RPC_CLIENT_KEY,
		},
		{
			name:      "trpc server",
			scopeName: "loongsuite.instrumentation.trpc",
			spanKind:  trace.SpanKindServer,
			expected:  utils.RPC_SERVER_KEY,
		},
		{
			name:      "kitex client",
			scopeName: "loongsuite.instrumentation.kitex",
			spanKind:  trace.SpanKindClient,
			expected:  utils.RPC_CLIENT_KEY,
		},
		{
			name:      "kitex server",
			scopeName: "loongsuite.instrumentation.kitex",
			spanKind:  trace.SpanKindServer,
			expected:  utils.RPC_SERVER_KEY,
		},
		{
			name:      "dubbo client",
			scopeName: "loongsuite.instrumentation.dubbo",
			spanKind:  trace.SpanKindClient,
			expected:  utils.RPC_CLIENT_KEY,
		},
		{
			name:      "dubbo server",
			scopeName: "loongsuite.instrumentation.dubbo",
			spanKind:  trace.SpanKindServer,
			expected:  utils.RPC_SERVER_KEY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getScopeKey(tt.scopeName, tt.spanKind)
			if result != tt.expected {
				t.Errorf("getScopeKey(%s, %v) = %v, want %v", tt.scopeName, tt.spanKind, result, tt.expected)
			}
		})
	}
}

func TestGetScopeKey_Database(t *testing.T) {
	tests := []struct {
		name      string
		scopeName string
		spanKind  trace.SpanKind
		expected  attribute.Key
	}{
		{
			name:      "mongo client",
			scopeName: "loongsuite.instrumentation.mongo",
			spanKind:  trace.SpanKindClient,
			expected:  utils.DB_CLIENT_KEY,
		},
		{
			name:      "goredisv9 client",
			scopeName: "loongsuite.instrumentation.goredisv9",
			spanKind:  trace.SpanKindClient,
			expected:  utils.DB_CLIENT_KEY,
		},
		{
			name:      "goredisv8 client",
			scopeName: "loongsuite.instrumentation.goredisv8",
			spanKind:  trace.SpanKindClient,
			expected:  utils.DB_CLIENT_KEY,
		},
		{
			name:      "gorm client",
			scopeName: "loongsuite.instrumentation.gorm",
			spanKind:  trace.SpanKindClient,
			expected:  utils.DB_CLIENT_KEY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getScopeKey(tt.scopeName, tt.spanKind)
			if result != tt.expected {
				t.Errorf("getScopeKey(%s, %v) = %v, want %v", tt.scopeName, tt.spanKind, result, tt.expected)
			}
		})
	}
}

func TestSpanKeySuppressor_UnifiedScopeName(t *testing.T) {
	// Test HTTP client span
	t.Run("HTTP client suppression", func(t *testing.T) {
		s := SpanKeySuppressor{
			spanKeys: []attribute.Key{utils.HTTP_CLIENT_KEY},
		}
		builder := Builder[testRequest, testResponse]{}
		builder.Init().SetSpanNameExtractor(testNameExtractor{}).
			SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
			SetInstrumentationScope(instrumentation.Scope{
				Name:    utils.NET_HTTP_CLIENT_SCOPE_NAME, // "loongsuite.instrumentation.nethttp"
				Version: "test",
			})
		instrumenter := builder.BuildInstrumenter()
		ctx := context.Background()
		traceProvider := sdktrace.NewTracerProvider()
		originalTP := otel.GetTracerProvider()
		otel.SetTracerProvider(traceProvider)
		defer otel.SetTracerProvider(originalTP)
		
		newCtx := instrumenter.Start(ctx, testRequest{})
		if !s.ShouldSuppress(newCtx, trace.SpanKindClient) {
			t.Errorf("should suppress HTTP client span")
		}
	})

	// Test HTTP server span
	t.Run("HTTP server suppression", func(t *testing.T) {
		s := SpanKeySuppressor{
			spanKeys: []attribute.Key{utils.HTTP_SERVER_KEY},
		}
		builder := Builder[testRequest, testResponse]{}
		builder.Init().SetSpanNameExtractor(testNameExtractor{}).
			SetSpanKindExtractor(&AlwaysServerExtractor[testRequest]{}).
			SetInstrumentationScope(instrumentation.Scope{
				Name:    utils.NET_HTTP_SERVER_SCOPE_NAME, // Same scope name as client
				Version: "test",
			})
		instrumenter := builder.BuildInstrumenter()
		ctx := context.Background()
		traceProvider := sdktrace.NewTracerProvider()
		originalTP := otel.GetTracerProvider()
		otel.SetTracerProvider(traceProvider)
		defer otel.SetTracerProvider(originalTP)
		
		newCtx := instrumenter.Start(ctx, testRequest{})
		if !s.ShouldSuppress(newCtx, trace.SpanKindServer) {
			t.Errorf("should suppress HTTP server span")
		}
	})

	t.Run("HTTP client doesn't suppress server", func(t *testing.T) {
		s := SpanKeySuppressor{
			spanKeys: []attribute.Key{utils.HTTP_SERVER_KEY},
		}
		builder := Builder[testRequest, testResponse]{}
		builder.Init().SetSpanNameExtractor(testNameExtractor{}).
			SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
			SetInstrumentationScope(instrumentation.Scope{
				Name:    utils.NET_HTTP_CLIENT_SCOPE_NAME,
				Version: "test",
			})
		instrumenter := builder.BuildInstrumenter()
		ctx := context.Background()
		traceProvider := sdktrace.NewTracerProvider()
		originalTP := otel.GetTracerProvider()
		otel.SetTracerProvider(traceProvider)
		defer otel.SetTracerProvider(originalTP)
		
		newCtx := instrumenter.Start(ctx, testRequest{})
		if s.ShouldSuppress(newCtx, trace.SpanKindClient) {
			t.Errorf("should not suppress - parent is client but looking for server key")
		}
	})
}

func TestSpanKindSuppressor_UnifiedScopeName(t *testing.T) {
	// Test same span kind suppression
	t.Run("Same kind suppression", func(t *testing.T) {
		sks := &SpanKindStrategy{}
		s := sks.create([]attribute.Key{})
		builder := Builder[testRequest, testResponse]{}
		builder.Init().SetSpanNameExtractor(testNameExtractor{}).
			SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
			SetInstrumentationScope(instrumentation.Scope{
				Name:    utils.GRPC_CLIENT_SCOPE_NAME, // "loongsuite.instrumentation.grpc"
				Version: "test",
			})
		instrumenter := builder.BuildInstrumenter()
		ctx := context.Background()
		traceProvider := sdktrace.NewTracerProvider()
		originalTP := otel.GetTracerProvider()
		otel.SetTracerProvider(traceProvider)
		defer otel.SetTracerProvider(originalTP)
		
		newCtx := instrumenter.Start(ctx, testRequest{})
		if !s.ShouldSuppress(newCtx, trace.SpanKindClient) {
			t.Errorf("should suppress same span kind")
		}
	})

	t.Run("Different kind no suppression", func(t *testing.T) {
		sks := &SpanKindStrategy{}
		s := sks.create([]attribute.Key{})
		builder := Builder[testRequest, testResponse]{}
		builder.Init().SetSpanNameExtractor(testNameExtractor{}).
			SetSpanKindExtractor(&AlwaysClientExtractor[testRequest]{}).
			SetInstrumentationScope(instrumentation.Scope{
				Name:    utils.GRPC_CLIENT_SCOPE_NAME, // Same scope name for client/server
				Version: "test",
			})
		instrumenter := builder.BuildInstrumenter()
		ctx := context.Background()
		traceProvider := sdktrace.NewTracerProvider()
		originalTP := otel.GetTracerProvider()
		otel.SetTracerProvider(traceProvider)
		defer otel.SetTracerProvider(originalTP)
		
		newCtx := instrumenter.Start(ctx, testRequest{})
		if s.ShouldSuppress(newCtx, trace.SpanKindServer) {
			t.Errorf("should not suppress different span kind")
		}
	})
}
