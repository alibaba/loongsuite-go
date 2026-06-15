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

type K8sEventSpanNameExtractor[REQUEST any, RESPONSE any] struct {
	Getter K8sEventAttrsGetter[REQUEST, RESPONSE]
}

func (e K8sEventSpanNameExtractor[REQUEST, RESPONSE]) Extract(request REQUEST) string {
	if e.Getter.GetK8sObjectKind(request) != "" {
		return "k8s.informer." + e.Getter.GetK8sObjectKind(request) + ".process"
	}
	return "k8s.informer.event.process"
}

type K8sEventsSpanNameExtractor[REQUEST any, RESPONSE any] struct {
	Getter K8sEventsAttrsGetter[REQUEST, RESPONSE]
}

func (e K8sEventsSpanNameExtractor[REQUEST, RESPONSE]) Extract(request REQUEST) string {
	if e.Getter.GetK8sEventsIsInInitialList(request) {
		return "k8s.informer.initial_list.process"
	}
	return "k8s.informer.events.process"
}
