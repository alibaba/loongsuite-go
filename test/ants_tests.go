// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

func init() {
	tc1 := NewGeneralTestCase("goants-v1-1.1.0-test", "goants-v1", "v1.1.0", "", "1.24", "", TestGoAntsV1)
	tc2 := NewGeneralTestCase("goants-v2-2.0.0-test", "goants-v2", "v2.0.0", "", "1.24", "", TestGoAntsV2)
	tc3 := NewMuzzleTestCase("goants-v1-muzzle-1.1.0-test", "github.com/panjf2000/ants", "goants-v1", "v1.1.0", "", "1.24", "", []string{"go", "build", "test_ants.go"})
	tc4 := NewMuzzleTestCase("goants-v2-muzzle-2.0.0-test", "github.com/panjf2000/ants/v2", "goants-v2", "v2.0.0", "", "1.24", "", []string{"go", "build", "test_ants.go"})

	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
	if tc3 != nil {
		TestCases = append(TestCases, tc3)
	}
	if tc4 != nil {
		TestCases = append(TestCases, tc4)
	}
}

func TestGoAntsV1(t *testing.T, env ...string) {
	UseApp("goants-v1/v1.1.0")
	RunGoBuild(t, "go", "build", "test_ants.go")
	RunApp(t, "test_ants", env...)
}

func TestGoAntsV2(t *testing.T, env ...string) {
	UseApp("goants-v2/v2.0.0")
	RunGoBuild(t, "go", "build", "test_ants.go")
	RunApp(t, "test_ants", env...)
}
