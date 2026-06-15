# 高级配置

# 1. 介绍
本指南详细介绍了如何有效地配置和使用otel工具。该工具允许您设置各种配置选项，构建您的项目，并自定义您的工作流程以获得最佳性能。

# 2. 命令
## `otel set`
配置该工具的主要方法是通过`otel set`命令。该命令允许您指定适合您需求的各种设置：

详细日志记录：启用详细日志记录以接收来自工具的详细输出，这有助于故障排除和理解工具的流程。
```bash
  $ otel set -verbose
```

调试模式：打开调试模式以收集调试级别的见解和信息。
```bash
  $ otel set -debug
```

多个配置：一次性设置多个配置。例如，在使用自定义规则文件的同时启用调试和详细模式：
```bash
  $ otel set -debug -verbose -rule=custom.json
```

仅自定义规则：禁用默认规则集，仅应用特定的自定义规则。当您需要为您的项目量身定制规则集时，这尤其有用。请注意，即使明确禁用，`base.json`也无法避免被启用。如果您对自定义规则感兴趣，请参阅[此文档](../dev/overview.md)。
```bash
  $ otel set -disable=all -rule=custom.json
```

禁用特定规则：禁用特定的默认规则，同时保持其他规则启用。这允许对应用的埋点规则进行细粒度控制。
```bash
  $ otel set -disable=gorm.json,redis.json
```

启用所有规则：启用所有规则。
```bash
  $ otel set -disable=
```

默认和自定义规则的组合：同时使用默认规则和自定义规则以提供全面的配置：
```bash
  $ otel set -rule=custom.json
```

多个规则文件：将多个自定义规则文件与默认规则结合使用，可以指定为逗号分隔的列表：
```bash
  $ otel set -rule=a.json,b.json
```

GoCache缓存路径：每次构建都会使用临时目录作为go缓存，您可以设置特定目录来重用缓存，这将减少编译时间超过50%。
```bash
  $ otel set -gocache=/tmp/go-cache
```

使用环境变量：除了使用`otel set`命令外，还可以使用环境变量覆盖配置。例如，`OTELTOOL_DEBUG`环境变量允许您暂时强制工具进入调试模式，使此方法对于一次性配置有效，而无需更改永久设置。

```bash
$ export OTELTOOL_DEBUG=true
$ export OTELTOOL_VERBOSE=true
$ export OTELTOOL_GO_CACHE="/tmp/go-otel-cache"
```

环境变量的名称对应于`otel set`命令中可用的配置选项，前缀为`OTELTOOL_`。

环境变量完整列表：

- `OTELTOOL_DEBUG`：启用调试模式。
- `OTELTOOL_VERBOSE`：启用详细日志记录。
- `OTELTOOL_RULE_JSON_FILES`：指定自定义规则文件。
- `OTELTOOL_DISABLE_RULES`：禁用特定规则。使用'all'禁用所有默认规则，或使用逗号分隔的规则文件名列表禁用特定规则。
- `OTELTOOL_GO_CACHE`：设置特定目录来重用go缓存。

这种方法为测试更改和试验配置提供了灵活性，而无需永久更改您现有的设置。

## `otel go build`
配置到位后，您可以使用带`otel`前缀的命令构建您的项目。这将工具的配置直接集成到构建过程中：

标准构建：使用默认设置构建您的项目。
```bash
  $ otel go build
```

输出到特定位置：构建您的项目并指定输出位置。
```bash
  $ otel go build -o app cmd/app
```

传递编译器标志：使用编译器标志进行更自定义的构建。
```bash
  $ otel go build -gcflags="-m" cmd/app
```
无论您的项目多么复杂，otel工具都通过自动为您的代码埋点以实现有效的可观察性来简化流程，唯一的要求是在您的构建命令中添加`otel`前缀。

## `otel version`

如果您想检查otel工具的版本，可以使用`otel version`命令。
```bash
  $ otel version
```

