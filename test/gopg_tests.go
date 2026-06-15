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
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

const gopg_dependency_name = "github.com/go-pg/pg/v10"
const gopg_module_name = "gopg"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("test_gopg_crud", gopg_module_name, "v10.10.0", "v10.14.0", "1.19", "", TestGopgCrud10140),
		NewLatestDepthTestCase("test_gopg_crud", gopg_dependency_name, gopg_module_name, "v10.10.0", "v10.14.0", "1.19", "", TestGopgCrud10140),
		NewGeneralTestCase("test_gopg_crud", gopg_module_name, "v10.10.0", "v10.14.0", "1.19", "", TestGopgCrud10100))
}

func TestGopgCrud10100(t *testing.T, env ...string) {
	_, postgresPort := initPostgresContainer()
	UseApp("gopg/v10.10.0")
	RunGoBuild(t, "go", "build", "test_gopg_crud.go")
	env = append(env, "POSTGRES_PORT="+postgresPort.Port())
	RunApp(t, "test_gopg_crud", env...)
}

func TestGopgCrud10140(t *testing.T, env ...string) {
	_, postgresPort := initPostgresContainer()
	UseApp("gopg/v10.14.0")
	RunGoBuild(t, "go", "build", "test_gopg_crud.go")
	env = append(env, "POSTGRES_PORT="+postgresPort.Port())
	RunApp(t, "test_gopg_crud", env...)
}

func initPostgresContainer() (testcontainers.Container, nat.Port) {
	containerReqeust := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections")}
	postgresC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: containerReqeust, Started: true})
	if err != nil {
		panic(err)
	}
	port, err := postgresC.MappedPort(context.Background(), "5432")
	if err != nil {
		panic(err)
	}
	return postgresC, port
}
