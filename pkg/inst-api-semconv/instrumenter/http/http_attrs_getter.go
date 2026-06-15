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

import "github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/net"

type HttpCommonAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetRequestMethod(request REQUEST) string
	GetHttpRequestHeader(request REQUEST, name string) []string
	GetHttpResponseStatusCode(request REQUEST, response RESPONSE, err error) int
	GetHttpResponseHeader(request REQUEST, response RESPONSE, name string) []string
	GetErrorType(request REQUEST, response RESPONSE, err error) string
}

type HttpServerAttrsGetter[REQUEST any, RESPONSE any] interface {
	HttpCommonAttrsGetter[REQUEST, RESPONSE]
	net.UrlAttrsGetter[REQUEST]
	net.NetworkAttrsGetter[REQUEST, RESPONSE]
	GetUrlScheme(request REQUEST) string
	GetUrlPath(request REQUEST) string
	GetUrlQuery(request REQUEST) string
	GetHttpRoute(request REQUEST) string
}

type HttpClientAttrsGetter[REQUEST any, RESPONSE any] interface {
	HttpCommonAttrsGetter[REQUEST, RESPONSE]
	net.NetworkAttrsGetter[REQUEST, RESPONSE]
	GetUrlFull(request REQUEST) string
	GetServerAddress(request REQUEST) string
	GetServerPort(request REQUEST) int
}
