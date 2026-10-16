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
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	semconvhttp "github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var netHttpServerInstrumenter = BuildNetHttpServerOtelInstrumenter()

//go:linkname serverOnEnter net/http.serverOnEnter
func serverOnEnter(call api.CallContext, _ interface{}, w http.ResponseWriter, r *http.Request) {
	if !netHttpEnabler.Enable() {
		return
	}
	if r == nil {
		return
	}
	if netHttpFilter.FilterUrl(r.URL) {
		return
	}
	request := &netHttpRequest{
		method:       r.Method,
		url:          r.URL,
		header:       r.Header,
		version:      getProtocolVersion(r.ProtoMajor, r.ProtoMinor),
		host:         r.Host,
		isTls:        r.TLS != nil,
	}
	request.requestHeaders = netHttpCaptureConfig.captureHeaders(r.Header)
	request.requestBody = captureHTTPRequestBody(r, netHttpCaptureConfig)
	ctx := netHttpServerInstrumenter.Start(r.Context(), request)
	if x, ok := call.GetParam(1).(http.ResponseWriter); ok {
		x1 := &writerWrapper{
			ResponseWriter: x,
			statusCode:     http.StatusOK,
			captureBody:    netHttpCaptureConfig.captureBody,
			maxBodyBytes:   netHttpCaptureConfig.maxBodyBytes,
		}
		call.SetParam(1, x1)
	}
	data := make(map[string]interface{}, 3)
	data["ctx"] = ctx
	data["request"] = request
	data["httpRequest"] = r
	call.SetData(data)
	return
}

//go:linkname serverOnExit net/http.serverOnExit
func serverOnExit(call api.CallContext) {
	if !netHttpEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request, ok := data["request"].(*netHttpRequest)
	if !ok {
		return
	}
	if rawRequest, ok := data["httpRequest"].(*http.Request); ok {
		span := trace.SpanFromContext(ctx)
		spanName := ""
		if readable, ok := span.(sdktrace.ReadOnlySpan); ok {
			spanName = readable.Name()
		}
		setServerRoutePattern(request, rawRequest, spanName)
		if route := (netHttpServerAttrsGetter{}).GetHttpRoute(request); route != "" &&
			(spanName == request.method ||
				(rawRequest.URL != nil && spanName == rawRequest.URL.Path)) {
			span.SetName(request.method + " " + route)
		}
	}
	if p, ok := call.GetParam(1).(http.ResponseWriter); ok {
		if w1, ok := p.(*writerWrapper); ok {
			netHttpServerInstrumenter.End(ctx, request, &netHttpResponse{
				statusCode:   w1.statusCode,
				header:       w1.Header(),
				hasResponse:  true,
				responseBody: w1.capturedResponseBody(),
			}, nil)
		}
	}

	return
}

func setServerRoutePattern(request *netHttpRequest, rawRequest *http.Request, spanName string) {
	// ServeMux fills Pattern after this hook's entry. Middleware that passes
	// r.WithContext(...) to the mux replaces the request pointer, so the outer
	// hook cannot observe that copy's pattern.
	request.routePattern = rawRequest.Pattern

	// Framework hooks publish validated route templates in the span name. They
	// are more specific than the outer ServeMux catch-all pattern (for example,
	// go-restful registers "/" while selecting "/users/{user-id}").
	if route := semconvhttp.RouteFromSpanName(spanName); route != "" {
		request.routePattern = route
		return
	}

	// A framework static route is also safe: it must exactly equal the request
	// path. Do not infer arbitrary request paths as route templates.
	if rawRequest.URL != nil && spanName == rawRequest.URL.Path {
		request.routePattern = spanName
	}
}

type writerWrapper struct {
	http.ResponseWriter
	statusCode           int
	captureBody          bool
	maxBodyBytes         int64
	responseBody         []byte
	responseBodyOverflow bool
}

func (w *writerWrapper) Write(data []byte) (int, error) {
	w.captureResponseBody(data)
	return w.ResponseWriter.Write(data)
}

func (w *writerWrapper) WriteHeader(statusCode int) {
	// cache the status code
	if w.statusCode == statusCode {
		return // Prevent multiple header writes
	}
	w.statusCode = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *writerWrapper) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not implement http.Hijacker")
}

func (w *writerWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *writerWrapper) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (w *writerWrapper) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
