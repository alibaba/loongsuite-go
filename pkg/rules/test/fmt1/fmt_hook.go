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

package fmt1

import (
	_ "fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname OnExitPrintf1 fmt.OnExitPrintf1
func OnExitPrintf1(call api.CallContext, n int, err error) {
	println("Exiting hook1....")
	call.SetReturnVal(0, 1024)
	v := call.GetData().(int)
	println(v)
}

//go:linkname OnEnterPrintf1 fmt.OnEnterPrintf1
func OnEnterPrintf1(call api.CallContext, format string, arg ...any) {
	println("Entering hook1....")
	call.SetData(555)
	call.SetParam(0, "olleH%s\n")
	p1 := call.GetParam(1).([]any)
	p1[0] = "goodcatch"
}
