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
	"os"
	"time"
)

type K8sClientGoInnerEnabler struct {
	enabled bool
}

var k8sEnabler = K8sClientGoInnerEnabler{
	enabled: os.Getenv("OTEL_INSTRUMENTATION_K8S_CLIENT_GO_ENABLED") != "false",
}

func (k K8sClientGoInnerEnabler) Enable() bool {
	return k.enabled
}

type k8sEventInfo struct {
	eventType       string
	eventUID        string
	namespace       string
	name            string
	resourceVersion string
	apiVersion      string
	kind            string
	startTime       time.Time
	processingTime  int64
}

type k8sEventsInfo struct {
	isInInitialList bool
	eventCount      int
	hasError        bool
	errorMsg        string
}
