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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"
)

var k8sClientGoEventsInstrumenter = BuildK8sClientGoEventsInstrumenter()

var _ k8s.K8sEventsAttrsGetter[k8sEventsInfo, k8sEventsInfo] = K8sEventsAttrsGetter{}

type K8sEventsAttrsGetter struct {
}

func (k K8sEventsAttrsGetter) GetK8sEventsIsInInitialList(info k8sEventsInfo) bool {
	return info.isInInitialList
}

func (k K8sEventsAttrsGetter) GetK8sEventsCount(info k8sEventsInfo) int {
	return info.eventCount
}

type K8sStatusExtractor[REQUEST k8sEventsInfo, RESPONSE k8sEventsInfo] struct{}

func (g K8sStatusExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request k8sEventsInfo, response k8sEventsInfo, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else if response.hasError {
		span.SetStatus(codes.Error, response.errorMsg)
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
}

func BuildK8sClientGoEventsInstrumenter() instrumenter.Instrumenter[k8sEventsInfo, k8sEventsInfo] {
	builder := instrumenter.Builder[k8sEventsInfo, k8sEventsInfo]{}
	getter := K8sEventsAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&k8s.K8sEventsSpanNameExtractor[k8sEventsInfo, k8sEventsInfo]{Getter: getter}).
		SetSpanKindExtractor(&instrumenter.AlwaysInternalExtractor[k8sEventsInfo]{}).
		SetSpanStatusExtractor(&K8sStatusExtractor[k8sEventsInfo, k8sEventsInfo]{}).
		AddAttributesExtractor(&k8s.K8sEventsAttrsExtractor[k8sEventsInfo, k8sEventsInfo, K8sEventsAttrsGetter]{Getter: getter}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.K8S_CLIENT_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
