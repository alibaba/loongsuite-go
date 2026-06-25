// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package newapi

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/gin-gonic/gin"

	_ "unsafe"
)

//go:linkname handleStreamFormatOnEnter github.com/QuantumNous/new-api/relay/channel/openai.handleStreamFormatOnEnter
func handleStreamFormatOnEnter(call api.CallContext, c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) {
	if !newAPIEnabler.Enable() {
		return
	}
	val, exists := c.Get(traceInfoCtxKey)
	if !exists {
		return
	}
	traceInfo, ok := val.(*streamTraceInfo)
	if !ok || traceInfo == nil {
		return
	}
	traceInfo.Messages = append(traceInfo.Messages, data)
}
