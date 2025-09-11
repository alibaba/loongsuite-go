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

const kratos_dependency_name = "github.com/go-kratos/kratos/v2"
const kratos_module_name = "kratos"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("kratos-http-test", kratos_module_name, "", "", "1.18", "", TestKratosHttp),
		NewGeneralTestCase("kratos-grpc-test", kratos_module_name, "", "", "1.18", "", TestKratosGrpc),
		NewLatestDepthTestCase("kratos-latest-depth-grpc", kratos_dependency_name, kratos_module_name, "v2.8.0", "", "1.22", "", TestKratosGrpc),
		NewLatestDepthTestCase("kratos-latest-depth-http", kratos_dependency_name, kratos_module_name, "v2.8.0", "", "1.22", "", TestKratosHttp),
		NewMuzzleTestCase("kratos-muzzle-grpc", kratos_dependency_name, kratos_module_name, "v2.8.0", "", "1.22", "", []string{"go", "build", "test_kratos_grpc.go", "server.go"}),
		NewMuzzleTestCase("kratos-muzzle-http", kratos_dependency_name, kratos_module_name, "v2.8.0", "", "1.22", "", []string{"go", "build", "test_kratos_http.go", "server.go"}),
	)
}

func TestKratosGrpc(t *testing.T, env ...string) {
	UseApp("kratos/v2.8.0")
	RunGoBuild(t, "go", "build", "test_kratos_grpc.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_grpc", env...)
}

func TestKratosHttp(t *testing.T, env ...string) {
	UseApp("kratos/v2.8.0")
	RunGoBuild(t, "go", "build", "test_kratos_http.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_http", env...)
}
