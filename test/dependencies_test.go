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

const DependenciesAppName = "dependencies"

func TestDependenciesMatch(t *testing.T) {
	UseApp(DependenciesAppName)

	// only test_dependencies.json and base.json should be available because -disable=all is set
	RunSet(t, "-disable=", "-verbose", "-rule=../../tool/data/test_dependencies.json")
	RunGoBuild(t, "go", "build", "main.go")

	stdout, _ := RunApp(t, "main")
	ExpectDebugLogContains(t, "Dependency github.com/missing/dependency not found for rule net/http")
	ExpectContains(t, stdout, "[DEP-TEST] nethttp Serve() called with dependency check")
	ExpectContains(t, stdout, "[DEP-TEST] nethttp serve() called with dependency check")
	ExpectNotContains(t, stdout, "[DEP-TEST] nethttp readRequest() This should NOT be instrumented due to missing dependency")
	ExpectNotContains(t, stdout, "[DEP-TEST] nethttp Write() This should NOT be instrumented due to missing dependency")

	t.Log("Successfully verified rules with matching dependencies are applied")
}
