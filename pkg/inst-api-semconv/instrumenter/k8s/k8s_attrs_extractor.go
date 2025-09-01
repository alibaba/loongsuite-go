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

package k8s

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

type K8sEventAttrsExtractor[REQUEST any, RESPONSE any, GETTER K8sEventAttrsGetter[REQUEST, RESPONSE]] struct {
	Getter GETTER
}

func (e K8sEventAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes,
		attribute.String("k8s.event.type", e.Getter.GetK8sEventType(request)),
		attribute.String("k8s.event.uid", e.Getter.GetK8sEventUID(request)),
		attribute.String("k8s.namespace.name", e.Getter.GetK8sNamespace(request)),
		attribute.String("k8s.object.name", e.Getter.GetK8sObjectName(request)),
		attribute.String("k8s.object.resource_version", e.Getter.GetK8sObjectResourceVersion(request)),
		attribute.String("k8s.object.api_version", e.Getter.GetK8sObjectAPIVersion(request)),
		attribute.String("k8s.object.kind", e.Getter.GetK8sObjectKind(request)),
		attribute.Int64("k8s.event.start_time", e.Getter.GetK8sEventStartTime(request)),
	)
	return attributes, parentContext
}

func (e K8sEventAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes,
		attribute.Int64("k8s.event.process_duration_ms", e.Getter.GetK8sEventProcessingTime(response)),
	)
	return attributes, context
}

type K8sEventsAttrsExtractor[REQUEST any, RESPONSE any, GETTER K8sEventsAttrsGetter[REQUEST, RESPONSE]] struct {
	Getter GETTER
}

func (e K8sEventsAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes,
		attribute.Bool("k8s.events.is_initial_list", e.Getter.GetK8sEventsIsInInitialList(request)),
		attribute.Int("k8s.events.count", e.Getter.GetK8sEventsCount(request)),
	)
	return attributes, parentContext
}

func (e K8sEventsAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, context context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, context
}
