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
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const es_v8_dependency_name = "github.com/elastic/go-elasticsearch/v8"
const es_v8_module_name = "elasticsearch"

const defaultHTTPPort = "9200"
const defaultTCPPort = "9300"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("es-crud-test", es_v8_module_name, "v8.4.0", "", "1.18", "", TestESCrud),
		NewGeneralTestCase("es-metrics-test", es_v8_module_name, "v8.4.0", "", "1.18", "", TestESMetrics),
		NewGeneralTestCase("es-typed-client-test", es_v8_module_name, "v8.4.0", "", "1.18", "", TestESTypedClient),
		NewGeneralTestCase("es-typed-client-metrics-test", es_v8_module_name, "v8.5.0", "", "1.18", "", TestEsTypedClientMetrics),
		NewLatestDepthTestCase("es-crud-latestdepth-test", es_v8_dependency_name, es_v8_module_name, "v8.4.0", "v8.15.0", "1.18", "", TestESCrud),
		NewMuzzleTestCase("es-muzzle", es_v8_dependency_name, es_v8_module_name, "v8.4.0", "v8.4.0", "1.18", "", []string{"go", "build", "test_es_crud.go"}),
		NewMuzzleTestCase("es-muzzle", es_v8_dependency_name, es_v8_module_name, "v8.5.0", "", "1.18", "", []string{"go", "build", "test_es_typedclient.go"}),
	)
}

func TestESCrud(t *testing.T, env ...string) {
	_, esPort := initElasticSearchContainer()
	UseApp("elasticsearch/v8.4.0")
	RunGoBuild(t, "go", "build", "test_es_crud.go")
	env = append(env, "OTEL_ES_PORT="+esPort.Port())
	RunApp(t, "test_es_crud", env...)
}

func TestESMetrics(t *testing.T, env ...string) {
	_, esPort := initElasticSearchContainer()
	UseApp("elasticsearch/v8.4.0")
	RunGoBuild(t, "go", "build", "test_es_metrics.go")
	env = append(env, "OTEL_ES_PORT="+esPort.Port())
	RunApp(t, "test_es_metrics", env...)
}

func TestESTypedClient(t *testing.T, env ...string) {
	_, esPort := initElasticSearchContainer()
	UseApp("elasticsearch/v8.5.0")
	RunGoBuild(t, "go", "build", "test_es_typedclient.go")
	env = append(env, "OTEL_ES_PORT="+esPort.Port())
	RunApp(t, "test_es_typedclient", env...)
}

func TestEsTypedClientMetrics(t *testing.T, env ...string) {
	_, esPort := initElasticSearchContainer()
	UseApp("elasticsearch/v8.5.0")
	RunGoBuild(t, "go", "build", "test_es_typedclient_metrics.go")
	env = append(env, "OTEL_ES_PORT="+esPort.Port())
	RunApp(t, "test_es_typedclient_metrics", env...)
}

func initElasticSearchContainer() (testcontainers.Container, nat.Port) {
	ctx := context.Background()
	elasticsearchContainer, err := runElasticSearchContainer(ctx)
	if err != nil {
		panic(err)
	}
	port, err := elasticsearchContainer.MappedPort(context.Background(), "9200")
	if err != nil {
		panic(err)
	}
	return elasticsearchContainer, port
}

func runElasticSearchContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.elastic.co/elasticsearch/elasticsearch:7.17.25",
			Env: map[string]string{
				"discovery.type": "single-node",
				"ES_JAVA_OPTS":   "-Xms512m -Xmx512m",
				"cluster.routing.allocation.disk.threshold_enabled": "false",
			},
			ExposedPorts: []string{
				defaultHTTPPort + "/tcp",
				defaultTCPPort + "/tcp",
			},
			WaitingFor: wait.ForLog("Cluster health status changed from [YELLOW] to [GREEN]"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, fmt.Errorf("generic container: %w", err)
	}

	return container, nil
}
