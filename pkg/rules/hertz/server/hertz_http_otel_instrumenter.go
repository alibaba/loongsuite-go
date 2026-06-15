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

package server

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"net/url"
	"strconv"

	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetRequest(req *protocol.Request) (dst *protocol.Request) {
	dst = &protocol.Request{}
	req.CopyToSkipBody(dst)
	return
}

type hertzHttpServerAttrsGetter struct {
}

func (n hertzHttpServerAttrsGetter) GetRequestMethod(request *protocol.Request) string {
	return string(request.Method())
}

func (n hertzHttpServerAttrsGetter) GetHttpRequestHeader(request *protocol.Request, name string) []string {
	keys := make([]string, 0)
	request.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (n hertzHttpServerAttrsGetter) GetHttpResponseStatusCode(request *protocol.Request, response *protocol.Response, err error) int {
	return response.StatusCode()
}

func (n hertzHttpServerAttrsGetter) GetHttpResponseHeader(request *protocol.Request, response *protocol.Response, name string) []string {
	keys := make([]string, 0)
	response.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (n hertzHttpServerAttrsGetter) GetErrorType(request *protocol.Request, response *protocol.Response, err error) string {
	return ""
}

func (n hertzHttpServerAttrsGetter) GetUrlScheme(request *protocol.Request) string {
	scheme := string(request.Scheme())
	if scheme != "" {
		return scheme
	}
	return "http"
}

func (n hertzHttpServerAttrsGetter) GetUrlPath(request *protocol.Request) string {
	return string(request.Path())
}

func (n hertzHttpServerAttrsGetter) GetUrlQuery(request *protocol.Request) string {
	return string(request.QueryString())
}

func (n hertzHttpServerAttrsGetter) GetNetworkType(request *protocol.Request, response *protocol.Response) string {
	return "ipv4"
}

func (n hertzHttpServerAttrsGetter) GetNetworkTransport(request *protocol.Request, response *protocol.Response) string {
	return "tcp"
}

func (n hertzHttpServerAttrsGetter) GetNetworkProtocolName(request *protocol.Request, response *protocol.Response) string {
	scheme := string(request.Scheme())
	if scheme != "" {
		return scheme
	}
	return "http"
}

func (n hertzHttpServerAttrsGetter) GetNetworkProtocolVersion(request *protocol.Request, response *protocol.Response) string {
	return ""
}

func (n hertzHttpServerAttrsGetter) GetNetworkLocalInetAddress(request *protocol.Request, response *protocol.Response) string {
	return ""
}

func (n hertzHttpServerAttrsGetter) GetNetworkLocalPort(request *protocol.Request, response *protocol.Response) int {
	return 0
}

func (n hertzHttpServerAttrsGetter) GetNetworkPeerInetAddress(request *protocol.Request, response *protocol.Response) string {
	return string(request.Host())
}

func (n hertzHttpServerAttrsGetter) GetNetworkPeerPort(request *protocol.Request, response *protocol.Response) int {
	return getPeerPort(request)
}

func (n hertzHttpServerAttrsGetter) GetHttpRoute(request *protocol.Request) string {
	return string(request.Path())
}

func getPeerPort(request *protocol.Request) int {
	u, err := url.Parse(GetRequest(request).URI().String())
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return 0
	}
	return port
}

type hertzTextMapCarrier struct {
	request *protocol.Request
}

func (h hertzTextMapCarrier) Get(key string) string {
	return h.request.Header.Get(key)
}

func (h hertzTextMapCarrier) Set(key string, value string) {
	h.request.SetHeader(key, value)
}

func (h hertzTextMapCarrier) Keys() []string {
	keys := make([]string, 0)
	h.request.Header.VisitAllCustomHeader(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func BuildHertzServerInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[*protocol.Request, *protocol.Response] {
	builder := instrumenter.Builder[*protocol.Request, *protocol.Response]{}
	serverGetter := hertzHttpServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpServerAttrsGetter, hertzHttpServerAttrsGetter]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpServerAttrsGetter]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpServerAttrsGetter]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[*protocol.Request, *protocol.Response]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[*protocol.Request, *protocol.Response]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[*protocol.Request]{}).
		AddOperationListeners(http.HttpServerMetrics("hertz.server")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.HERTZ_HTTP_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&http.HttpServerAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpServerAttrsGetter, hertzHttpServerAttrsGetter, hertzHttpServerAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor}).BuildPropagatingFromUpstreamInstrumenter(func(n *protocol.Request) propagation.TextMapCarrier {
		return hertzTextMapCarrier{n}
	}, otel.GetTextMapPropagator())
}
