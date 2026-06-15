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

type K8sEventAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetK8sNamespace(request REQUEST) string
	GetK8sObjectName(request REQUEST) string
	GetK8sObjectResourceVersion(request REQUEST) string
	GetK8sObjectAPIVersion(request REQUEST) string
	GetK8sObjectKind(request REQUEST) string
	GetK8sEventType(request REQUEST) string
	GetK8sEventUID(request REQUEST) string
	GetK8sEventProcessingTime(response RESPONSE) int64
	GetK8sEventStartTime(request REQUEST) int64
}

type K8sEventsAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetK8sEventsIsInInitialList(request REQUEST) bool
	GetK8sEventsCount(request REQUEST) int
}
