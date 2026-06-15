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

package fmt7

import (
	_ "fmt"
	_ "unsafe"

	aliasapi "github.com/alibaba/loongsuite-go-agent/pkg/api" // both alias api and instrumented package(fmt) are imported
)

//go:linkname onEnterSprintf3 fmt.onEnterSprintf3
func onEnterSprintf3(call aliasapi.CallContext, format string, arg ...any) {
	println("a3")
}

//go:linkname onExitSprintf3 fmt.onExitSprintf3
func onExitSprintf3(call aliasapi.CallContext, s string) {
	print("b3")
}
