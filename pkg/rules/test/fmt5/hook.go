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

package fmt5

import (
	_ "fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname OnEnterPrintf2 fmt.OnEnterPrintf2
func OnEnterPrintf2(call api.CallContext, format interface{}, arg ...interface{}) {
	println("hook2")
	for i := 0; i < 10; i++ {
		if i == 5 {
			panic("deliberately")
		}
	}
}
