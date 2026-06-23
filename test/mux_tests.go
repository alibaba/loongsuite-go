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

package test

import "testing"

const mux_dependency_name = "github.com/gorilla/mux"
const mux_module_name = "mux"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("mux-basic-test", mux_module_name, "v1.3.0", "", "1.18", "", TestBasicMux),
		NewGeneralTestCase("mux-middleware-test", mux_module_name, "v1.3.0", "", "1.18", "", TestMuxMiddleware),
		NewGeneralTestCase("mux-pattern-test", mux_module_name, "v1.3.0", "", "1.18", "", TestMuxPattern),
		NewGeneralTestCase("mux-prefix-test", mux_module_name, "v1.7.4", "", "1.18", "", TestMuxPrefix),
		NewMuzzleTestCase("mux-muzzle-test", mux_dependency_name, mux_module_name, "v1.3.0", "v1.6.2", "1.18", "", []string{"go", "build", "test_mux_basic.go"}),
		NewMuzzleTestCase("mux-muzzle-test", mux_dependency_name, mux_module_name, "v1.7.4", "", "1.18", "", []string{"go", "build", "test_mux_middleware.go"}),
		NewLatestDepthTestCase("mux-latestdepth-test", mux_dependency_name, mux_module_name, "v1.3.0", "", "1.18", "", TestBasicMux),
		NewLatestDepthTestCase("mux-middleware-latestdepth-test", mux_dependency_name, mux_module_name, "v1.7.4", "", "1.18", "", TestMuxMiddleware),
		NewLatestDepthTestCase("mux-pattern-latestdepth-test", mux_dependency_name, mux_module_name, "v1.3.0", "", "1.18", "", TestMuxPattern),
		NewLatestDepthTestCase("mux-prefix-latestdepth-test", mux_dependency_name, mux_module_name, "v1.7.4", "", "1.18", "", TestMuxPrefix),
	)
}

func TestBasicMux(t *testing.T, env ...string) {
	UseApp("mux/v1.3.0")
	RunGoBuild(t, "go", "build", "test_mux_basic.go")
	RunApp(t, "test_mux_basic", env...)
}

func TestMuxMiddleware(t *testing.T, env ...string) {
	UseApp("mux/v1.7.4")
	RunGoBuild(t, "go", "build", "test_mux_middleware.go")
	RunApp(t, "test_mux_middleware", env...)
}

func TestMuxPattern(t *testing.T, env ...string) {
	UseApp("mux/v1.3.0")
	RunGoBuild(t, "go", "build", "test_mux_pattern.go")
	RunApp(t, "test_mux_pattern", env...)
}
func TestMuxPrefix(t *testing.T, env ...string) {
	UseApp("mux/v1.7.4")
	RunGoBuild(t, "go", "build", "test_mux_prefix.go")
	RunApp(t, "test_mux_prefix", env...)
}
