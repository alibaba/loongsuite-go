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

package gomicro

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go-micro.dev/v5/metadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type GoMicroServerAttrsGetter struct {
}

func (n GoMicroServerAttrsGetter) GetRequestMethod(request goMicroServerRequest) string {
	return request.request.Method()
}
func (n GoMicroServerAttrsGetter) GetHttpRequestHeader(request goMicroServerRequest, name string) []string {
	all := make([]string, 0)
	md, ok := metadata.FromContext(request.ctx)
	if ok {
		value, ok := md.Get(name)
		if ok {
			all = append(all, string(value))
		}
	}
	return all
}
func (n GoMicroServerAttrsGetter) GetHttpResponseStatusCode(request goMicroServerRequest, response goMicroResponse, err error) int {
	if err != nil {
		return 500
	}
	return 200
}
func (n GoMicroServerAttrsGetter) GetHttpResponseHeader(request goMicroServerRequest, response goMicroResponse, name string) []string {
	all := make([]string, 0)
	md, ok := metadata.FromContext(response.ctx)
	if ok {
		value, ok := md.Get(name)
		if ok {
			all = append(all, string(value))
		}
	}
	return all
}
func (n GoMicroServerAttrsGetter) GetErrorType(request goMicroServerRequest, response goMicroResponse, err error) string {
	return ""
}
func (n GoMicroServerAttrsGetter) GetUrlScheme(request goMicroServerRequest) string {
	return "http"
}
func (n GoMicroServerAttrsGetter) GetUrlPath(request goMicroServerRequest) string {
	return request.request.Endpoint()
}
func (n GoMicroServerAttrsGetter) GetUrlQuery(request goMicroServerRequest) string {
	return ""
}
func (n GoMicroServerAttrsGetter) GetNetworkType(request goMicroServerRequest, response goMicroResponse) string {
	return "ipv4"
}
func (n GoMicroServerAttrsGetter) GetNetworkTransport(request goMicroServerRequest, response goMicroResponse) string {
	return "tcp"
}
func (n GoMicroServerAttrsGetter) GetNetworkProtocolName(request goMicroServerRequest, response goMicroResponse) string {
	return "http"
}
func (n GoMicroServerAttrsGetter) GetNetworkProtocolVersion(request goMicroServerRequest, response goMicroResponse) string {
	return ""
}
func (n GoMicroServerAttrsGetter) GetNetworkLocalInetAddress(request goMicroServerRequest, response goMicroResponse) string {
	return ""
}
func (n GoMicroServerAttrsGetter) GetNetworkLocalPort(request goMicroServerRequest, response goMicroResponse) int {
	return 0
}
func (n GoMicroServerAttrsGetter) GetNetworkPeerInetAddress(request goMicroServerRequest, response goMicroResponse) string {
	return request.request.Service()
}
func (n GoMicroServerAttrsGetter) GetNetworkPeerPort(request goMicroServerRequest, response goMicroResponse) int {
	return 0
}
func (n GoMicroServerAttrsGetter) GetHttpRoute(request goMicroServerRequest) string {
	return request.request.Endpoint()
}

type goMicroServerTextMapCarrier struct {
	request *goMicroServerRequest
}

func (h goMicroServerTextMapCarrier) Get(key string) string {
	mda, _ := metadata.FromContext(h.request.ctx)
	md := metadata.Copy(mda)
	value, _ := md.Get(key)
	return value
}

func (h goMicroServerTextMapCarrier) Set(key string, value string) {
	mda, _ := metadata.FromContext(h.request.ctx)
	md := metadata.Copy(mda)
	md.Set(key, value)
	h.request.ctx = metadata.NewContext(h.request.ctx, md)
}

func (h goMicroServerTextMapCarrier) Keys() []string {
	keys := make([]string, 0)
	mda, _ := metadata.FromContext(h.request.ctx)
	md := metadata.Copy(mda)
	for k, _ := range md {
		keys = append(keys, k)
	}
	return keys
}

func BuildGoMicroServerOtelInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[goMicroServerRequest, goMicroResponse] {
	builder := instrumenter.Builder[goMicroServerRequest, goMicroResponse]{}
	serverGetter := GoMicroServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[goMicroServerRequest, goMicroResponse, GoMicroServerAttrsGetter, GoMicroServerAttrsGetter]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[goMicroServerRequest, goMicroResponse, GoMicroServerAttrsGetter]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[goMicroServerRequest, goMicroResponse, GoMicroServerAttrsGetter]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[goMicroServerRequest, goMicroResponse]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[goMicroServerRequest, goMicroResponse]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[goMicroServerRequest]{}).
		AddOperationListeners(http.HttpServerMetrics("gomicro.server")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GOMICRO_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&http.HttpServerAttrsExtractor[goMicroServerRequest, goMicroResponse, GoMicroServerAttrsGetter, GoMicroServerAttrsGetter, GoMicroServerAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor}).BuildPropagatingFromUpstreamInstrumenter(func(n goMicroServerRequest) propagation.TextMapCarrier {
		return goMicroServerTextMapCarrier{request: &n}
	}, otel.GetTextMapPropagator())
}
