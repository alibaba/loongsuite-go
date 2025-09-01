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
	"testing"
)

const sqlx_dependency_name = "github.com/jmoiron/sqlx"
const sqlx_module_name = "sqlx"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("test_sqlx_crud", sqlx_module_name, "1.3.0", "v1.4.0", "1.19", "", TestSqlxCrudV130),
		NewLatestDepthTestCase("test_sqlx_latestdepth_crud", sqlx_dependency_name, sqlx_module_name, "1.3.0", "v1.4.0", "1.19", "", TestSqlxCrudV140),
		NewGeneralTestCase("test_sqlx_crud", sqlx_module_name, "1.3.0", "v1.4.0", "1.19", "", TestSqlxCrudV130))
}

func TestSqlxCrudV130(t *testing.T, env ...string) {
	_, mysqlPort := init5xMySqlContainer()
	UseApp("sqlx/v1.3.0")
	RunGoBuild(t, "go", "build", "test_sqlx_crud.go")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	RunApp(t, "test_sqlx_crud", env...)
}

func TestSqlxCrudV140(t *testing.T, env ...string) {
	_, mysqlPort := init5xMySqlContainer()
	UseApp("sqlx/v1.4.0")
	RunGoBuild(t, "go", "build", "test_sqlx_crud.go")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	RunApp(t, "test_sqlx_crud", env...)
}
