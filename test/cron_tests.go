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

const cronDependencyName = "github.com/robfig/cron/v3"
const cronModuleName = "cron"

func init() {
	tc1 := NewGeneralTestCase("cron-3.0.0-test", cronModuleName, "v3.0.0", "", "1.24", "", TestCronBasic)
	tc2 := NewMuzzleTestCase("cron-muzzle-3.0.0-test", cronDependencyName, cronModuleName, "v3.0.0", "", "1.24", "", []string{"go", "build", "test_cron.go"})

	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
}

func TestCronBasic(t *testing.T, env ...string) {
	UseApp("cron/v3.0.0")
	RunGoBuild(t, "go", "build", "test_cron.go")
	RunApp(t, "test_cron", env...)
}
