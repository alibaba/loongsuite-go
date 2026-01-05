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
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

const clickhousev2_dependency_name = "github.com/ClickHouse/clickhouse-go/v2"
const clickhousev2_module_name = "clickhousev2"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("test_clickhousev2_crud", clickhousev2_module_name, "v2.13.0", "v2.42.0", "1.23", "1.24", TestClickhousev2CrudV2130),
		NewLatestDepthTestCase("test_clickhousev2_latestdepth_crud", clickhousev2_dependency_name, clickhousev2_module_name, "v2.13.0", "v2.42.0", "1.23", "1.24", TestClickhousev2CrudV2420),
		NewGeneralTestCase("test_clickhousev2_crud", clickhousev2_module_name, "v2.13.0", "v2.42.0", "1.23", "1.24", TestClickhousev2CrudV2130))
}

func TestClickhousev2CrudV2130(t *testing.T, env ...string) {
	_, clickhousePort := initClickhouseContainer()
	UseApp("clickhousev2/v2.13.0")
	RunGoBuild(t, "go", "build", "test_clickhousev2_crud.go")
	env = append(env, "CLICKHOUSE_PORT="+clickhousePort.Port())
	RunApp(t, "test_clickhousev2_crud", env...)
}

func TestClickhousev2CrudV2420(t *testing.T, env ...string) {
	_, clickhousePort := initClickhouseContainer()
	UseApp("clickhousev2/v2.42.0")
	RunGoBuild(t, "go", "build", "test_clickhousev2_crud.go")
	env = append(env, "CLICKHOUSE_PORT="+clickhousePort.Port())
	RunApp(t, "test_clickhousev2_latestdepth_crud", env...)
}

func initClickhouseContainer() (testcontainers.Container, nat.Port) {
	containerReqeust := testcontainers.ContainerRequest{
		Image:        "docker.io/clickhouse/clickhouse-server:24.3.4.147",
		ExposedPorts: []string{"8123/tcp", "9000/tcp"},
		WaitingFor:   wait.ForListeningPort("9000/tcp").WithStartupTimeout(time.Minute)}
	clickhouseC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: containerReqeust, Started: true})
	if err != nil {
		panic(err)
	}
	port, err := clickhouseC.MappedPort(context.Background(), "9000")
	if err != nil {
		panic(err)
	}
	return clickhouseC, port
}
