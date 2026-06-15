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

package service

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strconv"
	"time"
	"unsafe"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/experimental"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_cache"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_grpc"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_http"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_error"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_request"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_response"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

//go:linkname beforeNamingHttpProxyCloseClient github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_http.beforeNamingHttpProxyCloseClient
func beforeNamingHttpProxyCloseClient(call api.CallContext, proxy *naming_http.NamingHttpProxy) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	t := reflect.ValueOf(proxy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return
	}
	beatReactorField := t.FieldByName("beatReactor")
	if beatReactorField.IsValid() && beatReactorField.CanInterface() {
		beatReactorInterface := beatReactorField.Interface()
		beatReactor, ok := beatReactorInterface.(naming_http.BeatReactor)
		if ok {
			if reg, ok := beatReactor.OtelReg.(metric.Registration); ok {
				err := reg.Unregister()
				if err != nil {
					log.Printf("failed to unregister metrics for beat reactor, %v", err)
				}
			}
		}
	}
}

//go:linkname beforeNamingClientClose github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client.beforeNamingClientClose
func beforeNamingClientClose(call api.CallContext, sc *naming_client.NamingClient) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	t := reflect.ValueOf(sc)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return
	}
	serviceInfoHolderField := t.FieldByName("serviceInfoHolder")
	if serviceInfoHolderField.IsValid() {
		holderPointer := serviceInfoHolderField.Pointer()
		holder := (*naming_cache.ServiceInfoHolder)(unsafe.Pointer(holderPointer))
		if holder != nil {
			if reg, ok := holder.OtelReg.(metric.Registration); ok {
				err := reg.Unregister()
				if err != nil {
					log.Printf("failed to unregister metrics for service info holder, %v", err)
				}
			}
		}
	}
}

//go:linkname beforeRequestToServer github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_grpc.beforeRequestToServer
func beforeRequestToServer(call api.CallContext, proxy *naming_grpc.NamingGrpcProxy, request rpc_request.IRequest) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	call.SetKeyData("ts", time.Now().UnixMilli())
	call.SetKeyData("request", request)
}

//go:linkname afterRequestToServer github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_grpc.afterRequestToServer
func afterRequestToServer(call api.CallContext, resp rpc_response.IResponse, err error) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	t := call.GetKeyData("ts").(int64)
	req := call.GetKeyData("request").(rpc_request.IRequest)
	code := "NA"
	if resp != nil {
		code = strconv.Itoa(resp.GetResultCode())
	}
	set := attribute.NewSet(attribute.KeyValue{
		Key:   "method",
		Value: attribute.StringValue("GRPC"),
	}, attribute.KeyValue{
		Key:   "type",
		Value: attribute.StringValue(req.GetRequestType()),
	}, attribute.KeyValue{
		Key:   "status",
		Value: attribute.StringValue(code),
	})
	experimental.ClientNamingRequestDuration.Record(context.Background(), float64(time.Now().UnixMilli()-t), metric.WithAttributeSet(set))
}

//go:linkname beforeCallServer github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server.beforeCallServer
func beforeCallServer(call api.CallContext, server *nacos_server.NacosServer, api string, params map[string]string, method string, curServer string, contextPath string) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	call.SetKeyData("ts", time.Now().UnixMilli())
	call.SetKeyData("method", method)
	call.SetKeyData("type", contextPath+api)
}

//go:linkname afterCallServer github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server.afterCallServer
func afterCallServer(call api.CallContext, result string, err error) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	method := call.GetKeyData("method").(string)
	t := call.GetKeyData("ts").(int64)
	tpe := call.GetKeyData("type").(string)
	code := "200"
	if err != nil {
		var nacosErr *nacos_error.NacosError
		errors.As(err, &nacosErr)
		code = nacosErr.ErrorCode()
	}
	set := attribute.NewSet(attribute.KeyValue{
		Key:   "method",
		Value: attribute.StringValue(method),
	}, attribute.KeyValue{
		Key:   "type",
		Value: attribute.StringValue(tpe),
	}, attribute.KeyValue{
		Key:   "status",
		Value: attribute.StringValue(code),
	})
	experimental.ClientNamingRequestDuration.Record(context.Background(), float64(time.Now().UnixMilli()-t), metric.WithAttributeSet(set))
}
