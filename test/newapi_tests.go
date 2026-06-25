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

const newapi_dependency_name = "github.com/QuantumNous/new-api"
const newapi_module_name = "newapi"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("newapi-text-helper-test", newapi_module_name,
			"", "", "1.22.0", "", TestNewAPITextHelper),
		NewMuzzleTestCase("newapi-muzzle-test", newapi_dependency_name, newapi_module_name,
			"", "", "1.22.0", "", []string{"go", "build", "test_text_helper.go"}),
	)
}

func TestNewAPITextHelper(t *testing.T, env ...string) {
	UseApp("newapi/v0.12.8")
	RunGoBuild(t, "go", "build", "test_text_helper.go")
	RunApp(t, "./test_text_helper", env...)
}
