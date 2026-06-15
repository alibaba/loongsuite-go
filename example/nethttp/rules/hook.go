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

package rules

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

//go:linkname httpClientEnterHook net/http.httpClientEnterHook
func httpClientEnterHook(call api.CallContext, t *http.Transport, req *http.Request) {
	header, _ := json.Marshal(req.Header)
	fmt.Println("request header is ", string(header))
}

//go:linkname httpClientExitHook net/http.httpClientExitHook
func httpClientExitHook(call api.CallContext, res *http.Response, err error) {
	// 添加对 res 是否为 nil 的检查
	if res != nil {
		header, _ := json.Marshal(res.Header)
		fmt.Println("response header is ", string(header))
	} else {
		fmt.Println("response is nil")
	}

	// 可以同时处理 err 信息
	if err != nil {
		fmt.Println("request error: ", err.Error())
	}
}
