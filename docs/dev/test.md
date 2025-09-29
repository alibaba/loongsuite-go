# Test the Hook Code

Once you've added a new instrumentation rule according
to [how-to-add-a-new-rule.md](https://github.com/alibaba/loongsuite-go-agent/blob/main/docs/how-to-add-a-new-rule.md),
you need to add tests to verify your rules. `loongsuite-go-agent` provides a convenient way to verify
your rules.

## Add a general plugin test case

Change directory to `/test`, and create a new directory for the plugin you want to test, for example `redis`. In
the `redis` directory, there are some subdirectories and the name of each subdirectory represents the lowest supported
version of the plugin. If you want to add tests for redis, you should do the following things:

### 1. Add lowest-version dependency for your rule

For example, if you add a rule that support redis from `v9.0.5` to the latest version, you should verify the lowest
redis version, which is `v9.0.5` first. You may create a subdirectory named `v9.0.5` and add the following `go.mod`:

```
module redis/v9.0.5

go 1.22

replace github.com/alibaba/loongsuite-go-agent => ../../../

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../../test/verifier

require (
	// import this dependency to use verifier
    github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.0.5
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/sdk v1.30.0
)
```

### 2. Write business logic based on the plugin

Then you need to write some business logic based on the `redis` plugin, for example, `test_executing_commands.go` does
the basic get and set operation in redis. Your tests should cover all the usage scenarios of this plugin as much as
possible.

### 3. Write verification code

Some telemetry data(like span) will be produced if the business code you've written is matched to the rule. For example,
there should be two spans produced by `test_executing_commands.go`, one represents for the `set` redis operation and the other
represents for the `get` redis operation. You should use the `verifier` to verify the correctness:

```go
import "github.com/alibaba/loongsuite-go-agent/test/verifier"

verifier.WaitAndAssertTraces(func (stubs []tracetest.SpanStubs) {
	verifier.VerifyDbAttributes(stubs[0][0], "set", "", "redis", "", "localhost", "set a b ex 5 ", "set", "")
	verifier.VerifyDbAttributes(stubs[1][0], "get", "", "redis", "", "localhost", "get a ", "get", "")
})
```

The `WaitAndAssertTraces` of the verifier accept a callback function, which provides all the traces that are produced.
You should verify the attribute, the parent context and all other key information of every span in all the traces.

If you want to verify the metrics data, you can also use the `verifier` like the following code:
```go
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics) {
		"http.server.request.duration": func(mrs metricdata.ResourceMetrics) {
		if len(mrs.ScopeMetrics) <= 0 {
			panic("No http.server.request.duration metrics received!")
		}
		point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
		if point.DataPoints[0].Count != 1 {
			panic("http.server.request.duration metrics count is not 1")
		}
		verifier.VerifyHttpServerMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "GET", "/a", "", "http", "1.1", "http", 200)
		},
		"http.client.request.duration": func(mrs metricdata.ResourceMetrics) {
		if len(mrs.ScopeMetrics) <= 0 {
			panic("No http.client.request.duration metrics received!")
		}
		point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
		if point.DataPoints[0].Count != 1 {
			panic("http.client.request.duration metrics count is not 1")
		}
		verifier.VerifyHttpClientMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "GET", "127.0.0.1:"+strconv.Itoa(port), "", "http", "1.1", port, 200)
       },
	})
```
Users need to use `WaitAndAssertMetrics` method in verifier to verify the correctness of the metrics data. `WaitAndAssertMetrics` receives a map,
the key of the map is the name of the metric, the value is the validation function for this metrics data. Users can write their own validation logic in the callback function.

### 4. Register tests

Finally, you should register the tests. You should write a `_tests.go` file in `test` directory to do the registration:

```go
const redis_dependency_name = "github.com/redis/go-redis/v9"
const redis_module_name = "redis"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("redis-9.0.5-executing-commands-test", redis_module_name, "v9.0.5", "v9.5.1", "1.18", "", TestExecutingCommands)
}

func TestExecutingCommands(t *testing.T, env ...string) {
	redisC, redisPort := initRedisContainer()
	defer clearRedisContainer(redisC)
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_executing_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_commands", env...)
}

```

In the `init` function, you need to wrap your test case with `NewGeneralTestCase`, which receives the
following arguments:

testName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion string, testFunc func(t *testing.T, env
...string)

1. testName: name of the test case.
2. moduleName: the subdirectory name in `test` directory.
3. minVersion: the lowest supported version of the plugin.
4. maxVersion: the highest supported version of the plugin
5. minGoVersion: the lowest supported Go version of the plugin.
6. maxGoVersion: the highest supported Go version of the plugin.
7. testFunc: test function to be executed.

You should build the test case with the `loongsuite-go-agent` to make your test case able to produce
telemetry data. Firstly you should call `UseApp` method to change directory to the directory of your test cases, and
then call `RunGoBuild` to do hybrid compilation. Finally, use the `RunApp` to run the instrumented test-case binary to
verify the telemetry data.

```go
func TestExecutingUnsupportedCommands(t *testing.T, env ...string) {
	redisC, redisPort := initRedisContainer()
	defer clearRedisContainer(redisC)
	UseApp("redis/v9.0.5")
	RunGoBuild(t, "go", "build", "test_executing_unsupported_commands.go")
	env = append(env, "REDIS_PORT="+redisPort.Port())
	RunApp(t, "test_executing_unsupported_commands", env...)
}
```

## Add a muzzle check case

Muzzle check is inspired
by [safety-mechanisms.md](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/docs/safety-mechanisms.md).
It is impossible for us to run general plugin test for every version because it's going to take a lot of time.
So `loongsuite-go-agent` will pick some random version to do the hybrid compilation in order to verify
the API compatibility between different versions. If the muzzle check finds that some APIs are changed in some version,
community will create a new rule to adapt it.

Users can add a muzzle check case by calling `NewMuzzleTestCase`, the arguments taken by `NewMuzzleTestCase` are almost
the same as `NewGeneralTestCase`. You need to additionally specify the dependency name of the plugin and the list of
classes that need to do the muzzle check.

## Add a latest-depth check case

Instrumentation tests are generally run against the lowest version of a library that we support to ensure a baseline
against users with old dependency versions. Due to the nature of the agent and locations where we instrument private
APIs, the agent may fail on a newly released version of the library. We run instrumentation tests additionally against
the latest version of the library, as fetched from remote, as part of a nightly build. If a new version of a library will
not work with the agent, we find out through this build and can address it by the next release of the agent.

Users can add a latest-depth check case by calling `NewLatestDepthTestCase`, the arguments taken by `NewLatestDepthTestCase`
are almost the same as `NewGeneralTestCase`. You need to additionally specify the dependency name of the plugin and the list of
classes that need to do the latest-depth check.

## Update the world test

World test is a comprehensive compatibility check designed to ensure that the Go agent correctly matches a wide range of 
plugin rules. It helps prevent missing instrumentation when adding or modifying rules, and ensures the rule system works across 
different versions of third-party libraries. The test verifies integrity by checking whether the number of matched ImportPath
values equals the expected count. If the counts do not match, all matched paths are logged for debugging.

Users can update the World test by modifying the `test/world_test.go` and `test/world/main.go` files. Add the relevant plugin import
path to test/world/main.go, and update the `expectImportCounts` variable in `test/world_test.go`. This ensures the completeness of 
rule matching.
