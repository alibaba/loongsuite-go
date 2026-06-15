// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nethttp1

import (
	"net/http"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname onEnterClientDo net/http.onEnterClientDo
func onEnterClientDo(call api.CallContext, recv *http.Client, req *http.Request) {
	println("Before Client.Do()")
}

//go:linkname onExitClientDo net/http.onExitClientDo
func onExitClientDo(call api.CallContext, resp *http.Response, err error) {
	panic("deliberately")
}
