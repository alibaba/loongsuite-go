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
	TestCases = append(TestCases, NewGeneralTestCase("test_clickhousev2_crud", clickhousev2_module_name, "2.3.0", "2.7.0", "1.19", "", TestClickhousev2CrudV230),
		NewLatestDepthTestCase("test_clickhousev2_latestdepth_crud", clickhousev2_dependency_name, clickhousev2_module_name, "1.3.0", "v1.7.0", "1.19", "", TestClickhousev2CrudV230),
		NewGeneralTestCase("test_clickhousev2_crud", clickhousev2_module_name, "1.3.0", "v1.7.0", "1.19", "", TestClickhousev2CrudV230))
}

func TestClickhousev2CrudV230(t *testing.T, env ...string) {
	//_, clickhousePort := initClickhouseContainer()
	UseApp("clickhousev2/v2.3.0")
	RunGoBuild(t, "go", "build", "test_clickhousev2_crud.go")
	//env = append(env, "CLICKHOUSE_PORT="+clickhousePort.Port())
	env = append(env, "CLICKHOUSE_PORT=8123")
	RunApp(t, "test_clickhousev2_crud", env...)
}

func initClickhouseContainer() (testcontainers.Container, nat.Port) {
	containerReqeust := testcontainers.ContainerRequest{
		Image:        "clickhouse:25.3",
		ExposedPorts: []string{"8123/tcp", "9000/tcp"},
		WaitingFor:   wait.ForLog("Startup complete").WithStartupTimeout(180 * time.Second)}
	cassandraC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: containerReqeust, Started: true})
	if err != nil {
		panic(err)
	}
	port, err := cassandraC.MappedPort(context.Background(), "8123")
	if err != nil {
		panic(err)
	}
	return cassandraC, port
}
