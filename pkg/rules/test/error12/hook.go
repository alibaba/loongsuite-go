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

package error12

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname onEnterTestGetSet errorstest/auxiliary.onEnterTestGetSet
func onEnterTestGetSet(call api.CallContext, arg1 int, arg2, arg3 bool, arg4 float64, arg5 string, arg6 interface{}, arg7, arg8 map[int]bool, arg9 chan int, arg10 []int) {
	call.SetParam(0, 7632)
	call.SetParam(1, arg2)
	call.SetParam(2, arg3)
	call.SetParam(3, arg4)
	call.SetParam(4, arg5)
	call.SetParam(5, arg6)
	call.SetParam(6, arg7)
	call.SetParam(7, arg8)
	call.SetParam(8, arg9)
	call.SetParam(9, arg10)
}

//go:linkname onExitTestGetSet errorstest/auxiliary.onExitTestGetSet
func onExitTestGetSet(call api.CallContext, arg1 int, arg2 bool, arg3 bool, arg4 float64, arg5 string, arg6 interface{}, arg7 map[int]bool, arg8 map[int]bool, arg9 chan int, arg10 []int) {
	call.SetReturnVal(0, arg1)
	call.SetReturnVal(1, arg2)
	call.SetReturnVal(2, arg3)
	call.SetReturnVal(3, arg4)
	call.SetReturnVal(4, arg5)
	call.SetReturnVal(5, arg6)
	call.SetReturnVal(6, arg7)
	call.SetReturnVal(7, arg8)
	call.SetReturnVal(8, arg9)
	call.SetReturnVal(9, arg10)
}
