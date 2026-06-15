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
	"testing"
)

const gorm_dependency_name = "gorm.io/gorm"
const gorm_module_name = "gorm"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("gorm_crud_test", gorm_module_name, "v1.23.0", "v1.24.6", "1.18", "", TestGormCrud1231),
		NewLatestDepthTestCase("gorm_latestdepth_test", gorm_dependency_name, gorm_module_name, "v1.23.0", "v1.24.6", "1.18", "", TestGormCrud1231),
		NewGeneralTestCase("gorm_crud_test", gorm_module_name, "v1.22.0", "v1.23.0", "1.18", "", TestGormCrud1220))
}

func TestGormCrud1231(t *testing.T, env ...string) {
	_, mysqlPort := init8xMySqlContainer()
	UseApp("gorm/v1.23.1")
	RunGoBuild(t, "go", "build", "test_gorm_crud.go")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	RunApp(t, "test_gorm_crud", env...)
}

func TestGormCrud1220(t *testing.T, env ...string) {
	_, mysqlPort := init8xMySqlContainer()
	UseApp("gorm/v1.22.0")
	RunGoBuild(t, "go", "build", "test_gorm_crud.go")
	env = append(env, "MYSQL_PORT="+mysqlPort.Port())
	RunApp(t, "test_gorm_crud", env...)
}
