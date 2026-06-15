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

package fmt4

import (
	_ "fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

type any = interface{}

//go:linkname onEnterSprintf1 fmt.onEnterSprintf1
func onEnterSprintf1(call api.CallContext, format string, arg ...any) {
	print("a1")
}

//go:linkname onExitSprintf1 fmt.onExitSprintf1
func onExitSprintf1(call api.CallContext, s string) {
	print("b1")
}
