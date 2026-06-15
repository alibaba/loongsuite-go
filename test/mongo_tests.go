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
	"github.com/testcontainers/testcontainers-go/wait"
)

const mongo_dependency_name = "go.mongodb.org/mongo-driver"
const mongo_module_name = "mongo"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("mongo-1.11.1-crud-test", mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", TestCrudMongo),
		NewGeneralTestCase("mongo-1.11.1-cursor-test", mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", TestCursor),
		NewGeneralTestCase("mongo-1.11.1-batch-test", mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", TestBatch),
		NewGeneralTestCase("mongo-1.11.1-metrics-test", mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", TestMetrics),
		NewMuzzleTestCase("mongo-1.11.1-crud-muzzle", mongo_dependency_name, mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", []string{"go", "build", "test_crud_mongo.go", "dsn.go"}),
		NewMuzzleTestCase("mongo-1.11.1-cursor-muzzle", mongo_dependency_name, mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", []string{"go", "build", "test_batch.go", "dsn.go"}),
		NewMuzzleTestCase("mongo-1.11.1-batch-muzzle", mongo_dependency_name, mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", []string{"go", "build", "test_cursor.go", "dsn.go"}),
		NewLatestDepthTestCase("mongo-1.11.1-latestDepth", mongo_dependency_name, mongo_module_name, "v1.11.1", "v1.15.1", "1.18", "", TestCrudMongo))
}

func TestCrudMongo(t *testing.T, env ...string) {
	_, mongoPort := initMongoContainer()
	UseApp("mongo/v1.11.1")
	RunGoBuild(t, "go", "build", "test_crud_mongo.go", "dsn.go")
	env = append(env, "MONGO_PORT="+mongoPort.Port())
	RunApp(t, "test_crud_mongo", env...)
}

func TestCursor(t *testing.T, env ...string) {
	_, mongoPort := initMongoContainer()
	UseApp("mongo/v1.11.1")
	RunGoBuild(t, "go", "build", "test_cursor.go", "dsn.go")
	env = append(env, "MONGO_PORT="+mongoPort.Port())
	RunApp(t, "test_cursor", env...)
}

func TestBatch(t *testing.T, env ...string) {
	_, mongoPort := initMongoContainer()
	UseApp("mongo/v1.11.1")
	RunGoBuild(t, "go", "build", "test_batch.go", "dsn.go")
	env = append(env, "MONGO_PORT="+mongoPort.Port())
	RunApp(t, "test_batch", env...)
}

func TestMetrics(t *testing.T, env ...string) {
	_, mongoPort := initMongoContainer()
	UseApp("mongo/v1.11.1")
	RunGoBuild(t, "go", "build", "test_metrics_mongo.go", "dsn.go")
	env = append(env, "MONGO_PORT="+mongoPort.Port())
	RunApp(t, "test_metrics_mongo", env...)
}

func initMongoContainer() (testcontainers.Container, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("waiting for connections"),
	}
	mongoC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	port, err := mongoC.MappedPort(context.Background(), "27017")
	if err != nil {
		panic(err)
	}
	return mongoC, port
}
