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

const grpc_dependency_name = "google.golang.org/grpc"
const grpc_module_name = "grpc"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("grpc-basic-test", grpc_module_name, "v1.77.0", "", "1.21", "", TestBasicGrpc),
		NewGeneralTestCase("grpc-metrics-test", grpc_module_name, "v1.77.0", "", "1.21", "", TestMetricsGrpc),
		NewGeneralTestCase("grpc-fail-status-test", grpc_module_name, "v1.77.0", "", "1.21", "", TestGrpcStatus),
		NewLatestDepthTestCase("grpc-latest-depth", grpc_dependency_name, grpc_module_name, "v1.77.0", "", "1.21", "", TestBasicGrpc),
		NewMuzzleTestCase("grpc-muzzle", grpc_dependency_name, grpc_module_name, "v1.77.0", "", "1.21", "", []string{"go", "build", "test_grpc_basic.go", "grpc_common.go", "grpc.pb.go", "grpc_grpc.pb.go"}),
		NewGeneralTestCase("grpc-client-stream-test", grpc_module_name, "v1.77.0", "", "1.21", "", TestGrpcClientStream),
	)
}

func TestBasicGrpc(t *testing.T, env ...string) {
	UseApp("grpc/v1.77.0")
	RunGoBuild(t, "go", "build", "test_grpc_basic.go", "grpc_common.go", "grpc.pb.go", "grpc_grpc.pb.go")
	env = append(env, "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn")
	RunApp(t, "test_grpc_basic", env...)
}

func TestGrpcStatus(t *testing.T, env ...string) {
	UseApp("grpc/v1.77.0")
	RunGoBuild(t, "go", "build", "test_grpc_fail_status.go", "grpc_common.go", "grpc.pb.go", "grpc_grpc.pb.go")
	env = append(env, "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn")
	RunApp(t, "test_grpc_fail_status", env...)
}

func TestGrpcClientStream(t *testing.T, env ...string) {
	UseApp("grpc/v1.77.0")
	RunGoBuild(t, "go", "build", "test_grpc_client_stream.go", "grpc_common.go", "grpc.pb.go", "grpc_grpc.pb.go")
	env = append(env, "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn")
	RunApp(t, "test_grpc_client_stream", env...)
}

func TestMetricsGrpc(t *testing.T, env ...string) {
	UseApp("grpc/v1.77.0")
	RunGoBuild(t, "go", "build", "test_grpc_metrics.go", "grpc_common.go", "grpc.pb.go", "grpc_grpc.pb.go")
	env = append(env, "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn")
	RunApp(t, "test_grpc_metrics", env...)
}
