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
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("databasesql-mysql-8x", "databasesql", "", "", "1.18", "", TestMySql8x),
		NewGeneralTestCase("databasesql-mysql-5x", "databasesql", "", "", "1.18", "", TestMySql5x),
	)
}

func TestMySql5x(t *testing.T, env ...string) {
	_, mysqlPort := init5xMySqlContainer()

	UseApp("databasesql/mysql")
	RunGoBuild(t, "go", "build")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	env = append(env, "OTEL_INSTRUMENTATION_DB_EXPERIMENTAL_ENABLE=true")
	RunApp(t, "mysql", env...)
}

func TestMySql8x(t *testing.T, env ...string) {
	_, mysqlPort := init8xMySqlContainer()

	UseApp("databasesql/mysql")
	RunGoBuild(t, "go", "build")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	env = append(env, "OTEL_INSTRUMENTATION_DB_EXPERIMENTAL_ENABLE=true")
	RunApp(t, "mysql", env...)
}

func init5xMySqlContainer() (testcontainers.Container, nat.Port) {
	ctx := context.Background()
	mysqlContainer, err := mysql.Run(ctx, "mysql:5.6")
	if err != nil {
		panic(err)
	}
	if err := mysqlContainer.Start(ctx); err != nil {
		panic(err)
	}
	port, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		panic(err)
	}
	return mysqlContainer, port
}

func init8xMySqlContainer() (testcontainers.Container, nat.Port) {
	ctx := context.Background()
	mysqlContainer, err := mysql.Run(ctx, "mysql:8.0.36")
	if err != nil {
		panic(err)
	}
	if err := mysqlContainer.Start(ctx); err != nil {
		panic(err)
	}
	port, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		panic(err)
	}
	return mysqlContainer, port
}
