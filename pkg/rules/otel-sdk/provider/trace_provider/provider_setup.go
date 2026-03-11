// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package span

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	_ "unsafe"
)

//go:linkname setTracerProviderOnEnter go.opentelemetry.io/otel.setTracerProviderOnEnter
func setTracerProviderOnEnter(call api.CallContext, tp trace.TracerProvider) {
	if otel.SetGlobalProviderEnable {
		call.SetSkipCall(true)
		return
	}
	otel.SetGlobalProviderEnable = true
}
