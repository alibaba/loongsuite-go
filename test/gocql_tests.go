// Copyright (c) 2025 Alibaba Group Holding Ltd.
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
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const gocql_dependency_name = "github.com/gocql/gocql"
const gocql_module_name = "gocql"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("test_gopg_crud", gocql_module_name, "1.3.0", "v1.7.0", "1.19", "", TestGocqlCrudV130),
		NewLatestDepthTestCase("test_gocql_latestdepth_crud", gocql_dependency_name, gocql_module_name, "1.3.0", "v1.7.0", "1.19", "", TestGocqlCrudV170),
		NewGeneralTestCase("test_gocql_crud", gocql_module_name, "1.3.0", "v1.7.0", "1.19", "", TestGocqlCrudV130))
}

func TestGocqlCrudV130(t *testing.T, env ...string) {
	_, cassandraPort := initCassandraContainer()
	UseApp("gocql/v1.3.0")
	RunGoBuild(t, "go", "build", "test_gocql_crud.go")
	env = append(env, "CASSANDRA_PORT="+cassandraPort.Port())
	RunApp(t, "test_gocql_crud", env...)
}

func TestGocqlCrudV170(t *testing.T, env ...string) {
	_, cassandraPort := initCassandraContainer()
	UseApp("gocql/v1.7.0")
	RunGoBuild(t, "go", "build", "test_gocql_crud.go")
	env = append(env, "CASSANDRA_PORT="+cassandraPort.Port())
	RunApp(t, "test_gocql_crud", env...)
}

func initCassandraContainer() (testcontainers.Container, nat.Port) {
	containerReqeust := testcontainers.ContainerRequest{
		Image:        "cassandra:4.1",
		ExposedPorts: []string{"9042/tcp"},
		Env: map[string]string{
			"JVM_OPTS":      "-Xms2G -Xmx2G",
			"MAX_HEAP_SIZE": "2G",
			"HEAP_NEWSIZE":  "800M",
		},
		WaitingFor: wait.ForLog("Startup complete").WithStartupTimeout(180 * time.Second)}
	cassandraC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: containerReqeust, Started: true})
	if err != nil {
		panic(err)
	}
	port, err := cassandraC.MappedPort(context.Background(), "9042")
	if err != nil {
		panic(err)
	}
	return cassandraC, port
}
