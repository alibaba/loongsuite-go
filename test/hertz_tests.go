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

const hertz_dependency_name = "github.com/cloudwego/hertz"
const hertz_module_name = "hertz"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("hertz-090-basic-test", hertz_module_name, "v0.9.0", "", "1.18", "1.24", TestBasicHertz),
		NewGeneralTestCase("hertz-090-basic-test-with-hook", hertz_module_name, "v0.9.0", "", "1.18", "1.24", TestBasicHertzWithHook),
		NewGeneralTestCase("hertz-090-basic-test-with-exception", hertz_module_name, "v0.9.0", "", "1.18", "1.24", TestBasicHertzWithException),
		NewGeneralTestCase("hertz-090-basic-test-with-regex", hertz_module_name, "v0.9.0", "", "1.18", "1.24", TestBasicHertzWithRegex),
		NewLatestDepthTestCase("hertz-090-basic-test-latestdepth", hertz_dependency_name, hertz_module_name, "v0.9.0", "", "1.18", "", TestBasicHertz),
		NewMuzzleTestCase("hertz-090-basic-muzzle", hertz_dependency_name, hertz_module_name, "v0.9.0", "v0.9.1", "1.18", "1.24", []string{"go", "build", "test_hertz_basic.go", "basic_func.go"}),
		NewMuzzleTestCase("hertz-090-basic-muzzle-high", hertz_dependency_name, hertz_module_name, "v0.9.1", "", "1.18", "", []string{"go", "build", "test_hertz_basic.go", "basic_func.go"}))
}

func TestBasicHertz(t *testing.T, env ...string) {
	UseApp("hertz/v0.10.1")
	RunGoBuild(t, "go", "build", "test_hertz_basic.go", "basic_func.go")
	RunApp(t, "test_hertz_basic", env...)
}

func TestBasicHertzWithHook(t *testing.T, env ...string) {
	UseApp("hertz/v0.10.1")
	RunGoBuild(t, "go", "build", "test_hertz_with_hook.go", "basic_func.go")
	RunApp(t, "test_hertz_with_hook", env...)
}

func TestBasicHertzWithException(t *testing.T, env ...string) {
	UseApp("hertz/v0.10.1")
	RunGoBuild(t, "go", "build", "test_hertz_with_exception.go", "basic_func.go")
	RunApp(t, "test_hertz_with_exception", env...)
}

func TestBasicHertzWithRegex(t *testing.T, env ...string) {
	UseApp("hertz/v0.10.1")
	RunGoBuild(t, "go", "build", "test_hertz_with_regex.go", "basic_func.go")
	RunApp(t, "test_hertz_with_regex", env...)
}
