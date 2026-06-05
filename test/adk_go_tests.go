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

import "testing"

const adk_go_dependency_name = "google.golang.org/adk"
const adk_go_module_name = "adk-go"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("adk-go-agent-test", adk_go_module_name, "v0.6.0", "", "1.24", "", TestADKGoAgent),
		NewMuzzleTestCase("adk-go-muzzle-test", adk_go_dependency_name, adk_go_module_name, "v0.6.0", "v0.9.99", "1.24", "", []string{"go", "build", "test_adk_agent.go"}),
		NewMuzzleTestCase("adk-go-v1-muzzle-test", adk_go_dependency_name, adk_go_module_name, "v1.0.0", "v1.0.0", "1.24", "", []string{"go", "build", "test_adk_agent.go"}),
		NewLatestDepthTestCase("adk-go-latest-depth-test", adk_go_dependency_name, adk_go_module_name, "v0.6.0", "v0.6.0", "1.24", "", TestADKGoAgent),
	)
}

func TestADKGoAgent(t *testing.T, env ...string) {
	UseApp("adk-go/v0.6.0")
	RunGoBuild(t, "go", "build", "test_adk_agent.go")
	RunApp(t, "test_adk_agent", env...)
}
