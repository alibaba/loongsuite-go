// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s_client_go

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"k8s.io/client-go/tools/cache"
)

//go:linkname ProcessDeltasOnEnter k8s.io/client-go/tools/cache.ProcessDeltasOnEnter
func ProcessDeltasOnEnter(call api.CallContext, handler cache.ResourceEventHandler, clientState cache.Store, deltas cache.Deltas, isInInitialList bool) {
	if !k8sEnabler.Enable() {
		return
	}
	eventsInfo := k8sEventsInfo{
		isInInitialList: isInInitialList,
		eventCount:      len(deltas),
	}
	ctx := k8sClientGoEventsInstrumenter.Start(context.TODO(), eventsInfo)

	handler = NewK8sOtelEventHandler(ctx, handler)
	call.SetParam(0, handler)

	m := map[string]interface{}{
		"ctx":        ctx,
		"eventsInfo": eventsInfo,
	}
	call.SetData(m)
}

//go:linkname ProcessDeltasOnExit k8s.io/client-go/tools/cache.ProcessDeltasOnExit
func ProcessDeltasOnExit(call api.CallContext, err error) {
	if !k8sEnabler.Enable() {
		return
	}
	m := call.GetData().(map[string]interface{})
	ctx := m["ctx"].(context.Context)
	eventsInfo := m["eventsInfo"].(k8sEventsInfo)
	if err != nil {
		eventsInfo.hasError = true
		eventsInfo.errorMsg = err.Error()
		k8sClientGoEventsInstrumenter.End(ctx, eventsInfo, eventsInfo, err)
	} else {
		k8sClientGoEventsInstrumenter.End(ctx, eventsInfo, eventsInfo, nil)
	}
}
