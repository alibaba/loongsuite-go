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

package nethttp3

import (
	"context"
	"io"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

// arg type has package prefix
//
//go:linkname onEnterNewRequestWithContext net/http.onEnterNewRequestWithContext
func onEnterNewRequestWithContext(call api.CallContext, ctx context.Context, method, url string, body io.Reader) {
	println("NewRequestWithContext()")
}
