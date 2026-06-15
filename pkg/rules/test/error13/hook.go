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

package error13

import (
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname onEnterTestGetSetRecv errorstest/auxiliary.onEnterTestGetSetRecv
func onEnterTestGetSetRecv(call api.CallContext, arg1 interface{}, arg2 int, arg3 float64) {
	recv := call.GetParam(0)
	v := reflect.ValueOf(recv)
	v = v.Elem()
	field := v.FieldByName("X")
	field.SetInt(4008208820)

	call.SetParam(1, 118888)
	call.SetParam(2, 0.001)
}

//go:linkname onExitTestGetSetRecv errorstest/auxiliary.onExitTestGetSetRecv
func onExitTestGetSetRecv(call api.CallContext, arg1 int, arg2 float64) {
	call.SetReturnVal(0, arg1)
	call.SetReturnVal(1, arg2)
}
