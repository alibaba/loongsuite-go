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

const nacos_dependency_name = "github.com/nacos-group/nacos-sdk-go/v2"
const nacos_module_name = "nacos"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("nacos-2.0.0-config-test", nacos_module_name, "v2.0.0", "v2.1.0", "1.18", "", TestNacos200Config),
		NewMuzzleTestCase("nacos-2.0.0-muzzle", nacos_dependency_name, nacos_module_name, "v2.0.0", "v2.1.0", "1.18", "", []string{"test_nacos_config.go"}),
		NewGeneralTestCase("nacos-2.0.0-service-test", nacos_module_name, "v2.0.0", "v2.1.0", "1.18", "", TestNacos200Service),
		NewGeneralTestCase("nacos-2.1.0-config-test", nacos_module_name, "v2.1.0", "", "1.18", "", TestNacos210Config),
		NewGeneralTestCase("nacos-2.1.0-service-test", nacos_module_name, "v2.1.0", "", "1.18", "", TestNacos210Service),
		NewMuzzleTestCase("nacos-2.1.0-muzzle", nacos_dependency_name, nacos_module_name, "v2.1.0", "", "1.18", "", []string{"test_nacos_config.go"}),
		NewLatestDepthTestCase("nacos-2.1.0-latestdepth-test", nacos_dependency_name, nacos_module_name, "", "v2.2.9", "1.18", "", TestNacos210Config))
}

func TestNacos200Config(t *testing.T, env ...string) {
	UseApp("nacos/v2.0.0")
	RunGoBuild(t, "go", "build", "test_nacos_config.go")
	env = append(env, "OTEL_INSTRUMENTATION_NACOS_EXPERIMENTAL_METRICS_ENABLE=true")
	RunApp(t, "test_nacos_config", env...)
}

func TestNacos200Service(t *testing.T, env ...string) {
	UseApp("nacos/v2.0.0")
	RunGoBuild(t, "go", "build", "test_nacos_service.go")
	env = append(env, "OTEL_INSTRUMENTATION_NACOS_EXPERIMENTAL_METRICS_ENABLE=true")
	RunApp(t, "test_nacos_service", env...)
}

func TestNacos210Config(t *testing.T, env ...string) {
	UseApp("nacos/v2.1.0")
	RunGoBuild(t, "go", "build", "test_nacos_config.go")
	env = append(env, "OTEL_INSTRUMENTATION_NACOS_EXPERIMENTAL_METRICS_ENABLE=true")
	RunApp(t, "test_nacos_config", env...)
}

func TestNacos210Service(t *testing.T, env ...string) {
	UseApp("nacos/v2.1.0")
	RunGoBuild(t, "go", "build", "test_nacos_service.go")
	env = append(env, "OTEL_INSTRUMENTATION_NACOS_EXPERIMENTAL_METRICS_ENABLE=true")
	RunApp(t, "test_nacos_service", env...)
}
