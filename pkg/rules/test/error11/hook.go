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

package error11

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname onEnterTestSkip2 errorstest/auxiliary.onEnterTestSkip2
func onEnterTestSkip2(call api.CallContext) {
	call.SetSkipCall(true)
}

//go:linkname onExitTestSkip2 errorstest/auxiliary.onExitTestSkip2
func onExitTestSkip2(call api.CallContext, _ int) {
	call.SetReturnVal(0, 0x512)
}
