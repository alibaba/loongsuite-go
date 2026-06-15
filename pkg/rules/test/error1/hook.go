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

package error1

import (
	_ "errors"
	"fmt"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname onEnterUnwrap errors.onEnterUnwrap
func onEnterUnwrap(call api.CallContext, err error) {
	newErr := fmt.Errorf("wrapped: %w", err)
	call.SetParam(0, newErr)
}

//go:linkname onExitUnwrap errors.onExitUnwrap
func onExitUnwrap(call api.CallContext, err error) {
	e := call.GetParam(0).(interface {
		Unwrap() error
	})
	old := e.Unwrap()
	fmt.Printf("old:%v\n", old)
}
