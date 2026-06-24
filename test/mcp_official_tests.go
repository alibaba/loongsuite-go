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

package test

import (
	"testing"
)

const mcp_official_dependency_name = "github.com/modelcontextprotocol/go-sdk"
const mcp_official_module_name = "mcp-official"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("mcp-official-v0.7.0-client-tool-call-test", mcp_official_module_name,
			"v0.7.0", "", "1.22.0", "", TestMcpOfficialClientToolCall),
		NewGeneralTestCase("mcp-official-v0.7.0-server-tool-call-test", mcp_official_module_name,
			"v0.7.0", "", "1.22.0", "", TestMcpOfficialServerToolCall),
	)
}

func TestMcpOfficialClientToolCall(t *testing.T, env ...string) {
	UseApp("mcp-official/v0.7.0")
	RunGoBuild(t, "go", "build", "test_client_tool_call.go", "ext.go")
	RunApp(t, "test_client_tool_call", env...)
}

func TestMcpOfficialServerToolCall(t *testing.T, env ...string) {
	UseApp("mcp-official/v0.7.0")
	RunGoBuild(t, "go", "build", "test_server_tool_call.go", "ext.go")
	RunApp(t, "test_server_tool_call", env...)
}
