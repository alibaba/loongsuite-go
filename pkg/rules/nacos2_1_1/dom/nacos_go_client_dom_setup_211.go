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

package dom

import (
	"context"
	"log"
	"reflect"
	"unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/experimental"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/cache"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_http"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

//go:linkname beforeNewBeatReactor211 github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_http.beforeNewBeatReactor211
func beforeNewBeatReactor211(call api.CallContext, ctx context.Context, clientCfg constant.ClientConfig, nacosServer *nacos_server.NacosServer) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	call.SetKeyData("namespace", clientCfg.NamespaceId)
	call.SetKeyData("region", clientCfg.RegionId)
	call.SetKeyData("appName", clientCfg.AppName)
	call.SetKeyData("appKey", clientCfg.AppKey)
	call.SetKeyData("userName", clientCfg.Username)
}

//go:linkname afterNewBeatReactor211 github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client/naming_http.afterNewBeatReactor211
func afterNewBeatReactor211(call api.CallContext, b naming_http.BeatReactor) {
	if !experimental.NacosEnabler.Enable() {
		return
	}
	t := reflect.ValueOf(&b).Elem()
	beatMapField := t.FieldByName("beatMap")
	if beatMapField.IsValid() {
		bf := reflect.NewAt(beatMapField.Type(), unsafe.Pointer(beatMapField.UnsafeAddr())).Elem()
		beatMap, ok := bf.Interface().(cache.ConcurrentMap)
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
			observer.ObserveInt64(experimental.ClientDomBeatMapSize, int64(beatMap.Count()), metric.WithAttributeSet(attrSet))
			return nil
		}, experimental.ClientDomBeatMapSize)
		if err != nil {
			log.Printf("[otel nacos] failed to register metrics for beat map, %v\n", err)
		} else {
			b.OtelReg = reg
			call.SetReturnVal(0, b)
		}
	}
}
