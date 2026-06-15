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

import (
	"testing"
)

const dubbo_dependency_name = "dubbo.apache.org/dubbo-go/v3"
const dubbo_module_name = "dubbo"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("dubbo-basic-test", dubbo_module_name, "v3.3.0", "", "1.21", "", TestBasicDubbo),
		NewGeneralTestCase("dubbo-metrics-test", dubbo_module_name, "v3.3.0", "", "1.21", "", TestMetricsDubbo),
		NewGeneralTestCase("dubbo-status-test", dubbo_module_name, "v3.3.0", "", "1.21", "", TestDubboStatus),
		NewLatestDepthTestCase("dubbo-latest-depth-test", dubbo_dependency_name, dubbo_module_name, "v3.3.0", "", "1.21", "", TestBasicDubbo),
		NewMuzzleTestCase("dubbo-muzzle-test-basic", dubbo_dependency_name, dubbo_module_name, "v3.3.0", "", "1.21", "", []string{"go", "build", "test_dubbo_basic.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go"}),
		NewMuzzleTestCase("dubbo-muzzle-test-matrics", dubbo_dependency_name, dubbo_module_name, "v3.3.0", "", "1.21", "", []string{"go", "build", "test_dubbo_metrics.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go"}),
		NewMuzzleTestCase("dubbo-muzzle-test-status", dubbo_dependency_name, dubbo_module_name, "v3.3.0", "", "1.21", "", []string{"go", "build", "test_dubbo_error.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go"}),
	)
}

func TestBasicDubbo(t *testing.T, env ...string) {
	UseApp("dubbo/v3.3.0")
	RunGoBuild(t, "go", "build", "test_dubbo_basic.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go")
	RunApp(t, "test_dubbo_basic", env...)
}

func TestMetricsDubbo(t *testing.T, env ...string) {
	UseApp("dubbo/v3.3.0")
	RunGoBuild(t, "go", "build", "test_dubbo_metrics.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go")
	RunApp(t, "test_dubbo_metrics", env...)
}

func TestDubboStatus(t *testing.T, env ...string) {
	UseApp("dubbo/v3.3.0")
	RunGoBuild(t, "go", "build", "test_dubbo_error.go", "dubbo_common.go", "greet.triple.go", "greet.pb.go")
	RunApp(t, "test_dubbo_error", env...)
}
