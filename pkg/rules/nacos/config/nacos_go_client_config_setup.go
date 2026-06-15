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

package config

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
	"github.com/nacos-group/nacos-sdk-go/v2/clients/cache"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_error"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_request"
	"github.com/nacos-group/nacos-sdk-go/v2/common/remote/rpc/rpc_response"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

//go:linkname beforeNewConfigClient github.com/nacos-group/nacos-sdk-go/v2/clients/config_client.beforeNewConfigClient
func beforeNewConfigClient(call api.CallContext, nc nacos_client.INacosClient) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	param, err := nc.GetClientConfig()
	if err != nil {
		return
	}
	call.SetKeyData("namespace", param.NamespaceId)
	call.SetKeyData("region", param.RegionId)
	call.SetKeyData("appName", param.AppName)
	call.SetKeyData("appKey", param.AppKey)
	call.SetKeyData("userName", param.Username)
}

//go:linkname afterNewConfigClient github.com/nacos-group/nacos-sdk-go/v2/clients/config_client.afterNewConfigClient
func afterNewConfigClient(call api.CallContext, client *config_client.ConfigClient, err error) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	if client == nil {
		return
	}
	// get reference for cache map
	t := reflect.ValueOf(client)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return
	}
	cacheMapField := t.FieldByName("cacheMap")
	if cacheMapField.IsValid() {
		cf := reflect.NewAt(cacheMapField.Type(), unsafe.Pointer(cacheMapField.UnsafeAddr())).Elem()
		cacheMap, ok := cf.Interface().(cache.ConcurrentMap)
		if !ok {
			return
		}
		attrSet := attribute.NewSet(attribute.KeyValue{
			Key:   "namespace",
			Value: attribute.StringValue(call.GetKeyData("namespace").(string)),
		}, attribute.KeyValue{
			Key:   "region",
			Value: attribute.StringValue(call.GetKeyData("region").(string)),
		}, attribute.KeyValue{
			Key:   "appName",
			Value: attribute.StringValue(call.GetKeyData("appName").(string)),
		}, attribute.KeyValue{
			Key:   "appKey",
			Value: attribute.StringValue(call.GetKeyData("appKey").(string)),
		}, attribute.KeyValue{
			Key:   "userName",
			Value: attribute.StringValue(call.GetKeyData("userName").(string)),
		})
		reg, err := experimental.GlobalMeter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
			observer.ObserveInt64(experimental.ClientConfigCacheMapSize, int64(cacheMap.Count()), metric.WithAttributeSet(attrSet))
			return nil
		}, experimental.ClientConfigCacheMapSize)
		if err != nil {
			log.Printf("[otel nacos] failed to register metrics for config info holder, %v\n", err)
		} else {
			client.OtelReg = reg
		}
	}
}

//go:linkname beforeConfigClientClose github.com/nacos-group/nacos-sdk-go/v2/clients/config_client.beforeConfigClientClose
func beforeConfigClientClose(call api.CallContext, sc *config_client.ConfigClient) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	if sc.OtelReg == nil {
		return
	}
	if reg, ok := sc.OtelReg.(metric.Registration); ok {
		err := reg.Unregister()
		if err != nil {
			log.Printf("[otel nacos] failed to unregister metrics for config info holder, %v", err)
		}
	}
}

//go:linkname beforeCallConfigServer github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server.beforeCallConfigServer
func beforeCallConfigServer(call api.CallContext, server *nacos_server.NacosServer, api string, params map[string]string, newHeaders map[string]string,
	method string, curServer string, contextPath string, timeoutMS uint64) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	call.SetKeyData("ts", time.Now().UnixMilli())
	call.SetKeyData("method", method)
	call.SetKeyData("type", contextPath+api)
}

//go:linkname afterCallConfigServer github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server.afterCallConfigServer
func afterCallConfigServer(call api.CallContext, result string, err error) {
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
	experimental.ClientConfigRequestDuration.Record(context.Background(), float64(time.Now().UnixMilli()-t), metric.WithAttributeSet(set))
}

//go:linkname beforeRequestProxy github.com/nacos-group/nacos-sdk-go/v2/clients/config_client.beforeRequestProxy
func beforeRequestProxy(call api.CallContext, cp *config_client.ConfigProxy, rpcClient *rpc.RpcClient, request rpc_request.IRequest, timeoutMills uint64) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	call.SetKeyData("ts", time.Now().UnixMilli())
	call.SetKeyData("request", request)
}

//go:linkname afterRequestProxy github.com/nacos-group/nacos-sdk-go/v2/clients/config_client.afterRequestProxy
func afterRequestProxy(call api.CallContext, resp rpc_response.IResponse, err error) {
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
	experimental.ClientConfigRequestDuration.Record(context.Background(), float64(time.Now().UnixMilli()-t), metric.WithAttributeSet(set))
}
