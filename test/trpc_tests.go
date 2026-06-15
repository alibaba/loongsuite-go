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
	"fmt"
	"testing"
)

const trpc_dependency_name = "trpc.group/trpc-go/trpc-go"
const trpc_module_name = "trpc"

var minSupportVersion, maxSupportVersion = "v1.0.0", ""
var minGoVersion, maxGoVersion = "1.22", ""

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("trpc-basic-test", trpc_module_name, minSupportVersion, maxSupportVersion, minGoVersion, maxGoVersion, TestBasicTrpc),
		NewGeneralTestCase("trpc-metrics-test", trpc_module_name, minSupportVersion, maxSupportVersion, minGoVersion, maxGoVersion, TestMetricsTrpc),
	)
}

func TestBasicTrpc(t *testing.T, env ...string) {
	UseApp(fmt.Sprintf("trpc/%s", minSupportVersion))
	RunGoBuild(t, "go", "build", "test_trpc_basic.go", "trpc_common.go", "hello.trpc.go", "hello.pb.go")
	RunApp(t, "test_trpc_basic", env...)
}

func TestMetricsTrpc(t *testing.T, env ...string) {
	UseApp(fmt.Sprintf("trpc/%s", minSupportVersion))
	RunGoBuild(t, "go", "build", "test_trpc_metrics.go", "trpc_common.go", "hello.trpc.go", "hello.pb.go")
	env = append(env, "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn")
	RunApp(t, "test_trpc_metrics", env...)
}
