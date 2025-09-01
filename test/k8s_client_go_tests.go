// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

const client_go_dependency_name = "k8s.io/client-go"
const client_go_module_name = "k8s-client-go"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("k8s-client-go-basic-test", client_go_module_name, "v0.33.3", "", "1.24", "", TestBasicK8sClientGo),
	)
}

func TestBasicK8sClientGo(t *testing.T, env ...string) {
	k3sContainer, kubeconfigYaml := initK3sContainer()
	defer testcontainers.CleanupContainer(t, k3sContainer)
	UseApp("k8s-client-go/v0.33.3")
	RunGoBuild(t, "go", "build", "test_k8s_basic.go", "k8s_common.go")
	env = append(env, "KUBECONFIG="+string(kubeconfigYaml))
	RunApp(t, "test_k8s_basic", env...)
}

func initK3sContainer() (*k3s.K3sContainer, string) {
	ctx := context.Background()
	k3sContainer, err := k3s.Run(ctx, "rancher/k3s:v1.27.1-k3s1")
	if err != nil {
		panic(err)
	}
	kubeconfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		panic(err)
	}
	return k3sContainer, string(kubeconfigYaml)
}
