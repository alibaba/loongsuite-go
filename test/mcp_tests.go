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

const mcp_dependency_name = "github.com/mark3labs/mcp-go/mcp"
const mcp_module_name = "mcp"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("mcp-0.20.0-sse-tool-test", mcp_module_name, "0.20.0", "0.20.0", "1.22.0", "", TestMcpTool),
		NewGeneralTestCase("mcp-0.20.0-sse-prompt-test", mcp_module_name, "0.20.0", "0.20.0", "1.22.0", "", TestMcpPrompt),
		NewGeneralTestCase("mcp-0.20.0-sse-resource-test", mcp_module_name, "0.20.0", "0.20.0", "1.22.0", "", TestMcpResource),
	)

}

func TestMcpTool(t *testing.T, env ...string) {
	UseApp("mcp/v0.20.0")
	RunGoBuild(t, "go", "build", "test_sse_tool.go", "ext.go")
	RunApp(t, "test_sse_tool", env...)
}
func TestMcpPrompt(t *testing.T, env ...string) {
	UseApp("mcp/v0.20.0")
	RunGoBuild(t, "go", "build", "test_sse_prompt.go", "ext.go")
	RunApp(t, "test_sse_prompt", env...)
}
func TestMcpResource(t *testing.T, env ...string) {
	UseApp("mcp/v0.20.0")
	RunGoBuild(t, "go", "build", "test_sse_resource.go", "ext.go")
	RunApp(t, "test_sse_resource", env...)
}

// Since standard input/output communication cannot be implemented in this test, it will hang the test process. Therefore, this method is only reserved for future use when stdio is available and is not used currently.
// Since standard input/output communication cannot be implemented in the test here and will cause the test process to hang, this method is reserved for future use when stdio becomes available, and is currently not used.
/*func TestStdioTool(t *testing.T, env ...string) {
	UseApp("mcp/v0.20.0")
	RunGoBuild(t, "go", "build", "test_stdio_server.go", "ext.go")
	fmt.Println("a1")
	RunGoBuild(t, "go", "build", "test_stdio_tool.go")
	fmt.Println("a2")
	RunApp(t, "test_stdio_tool", env...)
}*/
