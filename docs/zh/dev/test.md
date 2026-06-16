# 测试Hook代码

根据[如何添加新规则.md](https://github.com/alibaba/loongsuite-go/blob/main/docs/how-to-add-a-new-rule.md)添加新的埋点规则后，您需要添加测试来验证您的规则。`loongsuite-go`提供了一种方便的方式来验证您的规则。

## 添加一个通用的插件测试用例

切换目录到`/test`，并为您想要测试的插件创建一个新目录，例如`redis`。在`redis`目录中，有一些子目录，每个子目录的名称代表插件支持的最低版本。如果您想为redis添加测试，您应该执行以下操作：

### 1. 为您的规则添加最低版本的依赖

例如，如果您添加一个支持从`v9.0.5`到最新版本的redis的规则，您应该首先验证最低的redis版本，即`v9.0.5`。您可以创建一个名为`v9.0.5`的子目录，并添加以下`go.mod`：

```
module redis/v9.0.5

go 1.22

replace github.com/alibaba/loongsuite-go => ../../../

replace github.com/alibaba/loongsuite-go/test/verifier => ../../../test/verifier

require (
	// 导入此依赖以使用验证器
    github.com/alibaba/loongsuite-go/test/verifier v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.0.5
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/sdk v1.30.0
)
```

### 2. 基于插件编写业务逻辑

然后，您需要基于`redis`插件编写一些业务逻辑，例如，`test_executing_commands.go`在redis中执行基本的get和set操作。您的测试应尽可能覆盖此插件的所有使用场景。

### 3. 编写验证代码

如果您编写的业务代码与规则匹配，则会产生一些遥测数据（如span）。例如，`test_executing_commands.go`应该产生两个span，一个代表`set` redis操作，另一个代表`get` redis操作。您应该使用`verifier`来验证其正确性：

```go
import "github.com/alibaba/loongsuite-go/test/verifier"

verifier.WaitAndAssertTraces(func (stubs []tracetest.SpanStubs) {
	verifier.VerifyDbAttributes(stubs[0][0], "set", "", "redis", "", "localhost", "set a b ex 5 ", "set", "")
	verifier.VerifyDbAttributes(stubs[1][0], "get", "", "redis", "", "localhost", "get a ", "get", "")
})
```

验证器的`WaitAndAssertTraces`接受一个回调函数，该函数提供所有产生的trace。您应该验证所有trace中每个span的属性、父上下文以及所有其他关键信息。

如果您想验证指标数据，您也可以使用`verifier`，如下面的代码所示：
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
用户需要使用verifier中的`WaitAndAssertMetrics`方法来验证指标数据的正确性。`WaitAndAssertMetrics`接收一个map，map的键是指标的名称，值是该指标数据的验证函数。用户可以在回调函数中编写自己的验证逻辑。

### 4. 注册测试

最后，您应该注册测试。您应该在`test`目录中编写一个`_tests.go`文件来进行注册：

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

在`init`函数中，您需要使用`NewGeneralTestCase`包装您的测试用例，它接收以下参数：

testName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion string, testFunc func(t *testing.T, env
...string)

1. testName：测试用例的名称。
2. moduleName：`test`目录中的子目录名称。
3. minVersion：插件支持的最低版本。
4. maxVersion：插件支持的最高版本。
5. minGoVersion：插件支持的最低Go版本。
6. maxGoVersion：插件支持的最高Go版本。
7. testFunc：要执行的测试函数。

您应该使用`loongsuite-go`构建测试用例，以使您的测试用例能够生成遥测数据。首先，您应该调用`UseApp`方法将目录更改为您的测试用例的目录，然后调用`RunGoBuild`进行混合编译。最后，使用`RunApp`运行已埋点的测试用例二进制文件以验证遥测数据。

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

## 添加一个muzzle检查用例

Muzzle检查的灵感来自于[safety-mechanisms.md](https://github.com/open-telemetry/opentelemetry-java-instrumentation/blob/main/docs/safety-mechanisms.md)。我们不可能为每个版本都运行通用的插件测试，因为这会花费大量时间。因此，`loongsuite-go`会选择一些随机版本进行混合编译，以验证不同版本之间的API兼容性。如果muzzle检查发现某些API在某个版本中发生了更改，社区将创建一个新规则来适应它。

用户可以通过调用`NewMuzzleTestCase`来添加一个muzzle检查用例，`NewMuzzleTestCase`所带的参数与`NewGeneralTestCase`几乎相同。您需要额外指定插件的依赖项名称以及需要进行muzzle检查的类列表。

## 添加一个latest-depth检查用例

埋点测试通常针对我们支持的库的最低版本运行，以确保与使用旧依赖版本的用户有一个基线。由于agent的性质以及我们埋点私有API的位置，agent可能会在库的新发布版本上失败。我们会在夜间构建中，额外针对从远程获取的最新版本的库运行埋点测试。如果库的新版本与agent不兼容，我们通过此构建发现并可以在agent的下一个版本中解决它。

用户可以通过调用`NewLatestDepthTestCase`来添加一个latest-depth检查用例，`NewLatestDepthTestCase`所带的参数与`NewGeneralTestCase`几乎相同。您需要额外指定插件的依赖项名称以及需要进行latest-depth检查的类列表。

## 更新world测试

World测试是一个全面的兼容性检查，旨在确保Go agent能够正确匹配各种插件规则。它有助于在添加或修改规则时防止遗漏埋点，并确保规则系统在不同版本的第三方库中正常工作。该测试通过检查匹配的ImportPath值的数量是否等于预期计数来验证完整性。如果计数不匹配，则会记录所有匹配的路径以进行调试。

用户可以通过修改`test/world_test.go`和`test/world/main.go`文件来更新World测试。将相关的插件导入路径添加到`test/world/main.go`，并更新`test/world_test.go`中的`expectImportCounts`变量。这确保了规则匹配的完整性。

