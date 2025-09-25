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

const kitex_dependency_name = "github.com/cloudwego/kitex"
const kitex_module_name = "kitex"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("kitex-basic-test", kitex_module_name, "", "", "1.18", "", TestKitexBasic),
		NewGeneralTestCase("kitex-grpc-test", kitex_module_name, "", "", "1.18", "", TestKitexGrpc),
		NewMuzzleTestCase("kitex-basic-test", kitex_dependency_name, kitex_module_name, "", "", "1.18", "", []string{"test_grpc_kitex.go", "handler.go"}),
		NewLatestDepthTestCase("kitex-latestdepth-test", kitex_dependency_name, kitex_module_name, "", "v0.11.3", "1.18", "", TestKitexBasic),
	)
}

func TestKitexBasic(t *testing.T, env ...string) {
	UseApp("kitex/v0.5.1")
	RunGoBuild(t, "go", "build", "test_basic_kitex.go", "handler.go")
	RunApp(t, "test_basic_kitex", env...)
}

func TestKitexGrpc(t *testing.T, env ...string) {
	UseApp("kitex/v0.5.1")
	RunGoBuild(t, "go", "build", "test_grpc_kitex.go", "handler.go")
	RunApp(t, "test_grpc_kitex", env...)
}
