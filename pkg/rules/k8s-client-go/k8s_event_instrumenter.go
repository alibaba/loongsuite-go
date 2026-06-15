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
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/k8s"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var k8sClientGoEventInstrumenter = BuildK8sClientGoEventInstrumenter()

var _ k8s.K8sEventAttrsGetter[k8sEventInfo, k8sEventInfo] = K8sEventAttrsGetter{}

type K8sEventAttrsGetter struct {
}

func (k K8sEventAttrsGetter) GetK8sNamespace(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.namespace
}

func (k K8sEventAttrsGetter) GetK8sObjectName(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.name
}

func (k K8sEventAttrsGetter) GetK8sObjectResourceVersion(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.resourceVersion
}

func (k K8sEventAttrsGetter) GetK8sObjectAPIVersion(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.apiVersion
}

func (k K8sEventAttrsGetter) GetK8sObjectKind(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.kind
}

func (k K8sEventAttrsGetter) GetK8sEventType(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.eventType
}

func (k K8sEventAttrsGetter) GetK8sEventUID(k8sEventInfo k8sEventInfo) string {
	return k8sEventInfo.eventUID
}

func (k K8sEventAttrsGetter) GetK8sEventProcessingTime(k8sEventInfo k8sEventInfo) int64 {
	return k8sEventInfo.processingTime
}

func (k K8sEventAttrsGetter) GetK8sEventStartTime(k8sEventInfo k8sEventInfo) int64 {
	return k8sEventInfo.startTime.UnixNano()
}

func BuildK8sClientGoEventInstrumenter() instrumenter.Instrumenter[k8sEventInfo, k8sEventInfo] {
	builder := instrumenter.Builder[k8sEventInfo, k8sEventInfo]{}
	getter := K8sEventAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&k8s.K8sEventSpanNameExtractor[k8sEventInfo, k8sEventInfo]{Getter: getter}).
		SetSpanKindExtractor(&instrumenter.AlwaysInternalExtractor[k8sEventInfo]{}).
		AddAttributesExtractor(&k8s.K8sEventAttrsExtractor[k8sEventInfo, k8sEventInfo, K8sEventAttrsGetter]{Getter: getter}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.K8S_CLIENT_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
