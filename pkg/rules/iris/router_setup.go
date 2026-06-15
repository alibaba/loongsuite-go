// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package iris

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	iContext "github.com/kataras/iris/v12/context"
	"go.opentelemetry.io/otel/sdk/trace"
)

//go:linkname irisHttpOnEnter github.com/kataras/iris/v12/core/router.irisHttpOnEnter
func irisHttpOnEnter(call api.CallContext, _ interface{}, iCtx *iContext.Context) {
	if !irisEnabler.Enable() {
		return
	}
	if iCtx == nil {
		return
	}
	r := iCtx.Request()
	lcs := trace.LocalRootSpanFromGLS()
	if lcs != nil && r != nil && iCtx.Path() != "" && r.URL != nil && (iCtx.Path() != r.URL.Path) {
		lcs.SetName(iCtx.Path())
	}
}
