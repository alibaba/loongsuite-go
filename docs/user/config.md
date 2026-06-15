# Advanced Configuration

# 1. Introduction
This guide provides a detailed overview of configuring and using the otel tool effectively. This tool allows you to set various configuration options, build your projects, and customize your workflow for optimal performance.

# 2. Commands
## `otel set`
The primary method of configuring the tool is through the `otel set` command. This command allows you to specify various settings tailored to your needs:

Verbose Logging: Enable verbose logging to receive detailed output from the tool, which is helpful for troubleshooting and understanding the tool's processes.
```bash
  $ otel set -verbose
```

Debug Mode: Turn on debug mode to gather debug-level insights and information.
```bash
  $ otel set -debug
```

Multiple Configurations: Set multiple configurations at once. For instance, enable both debug and verbose modes while using a custom rule file:
```bash
  $ otel set -debug -verbose -rule=custom.json
```

Custom Rules Only: Disable the default rule set and apply only specific custom rules. This is particularly useful when you need a tailored rule set for your project. Note that `base.json` is inevitable to be enabled, even if it is explicitly disabled. If you are interested in custom rules, please refer to [this document](../dev/overview.md).
```bash
  $ otel set -disable=all -rule=custom.json
```

Disable Specific Rules: Disable specific default rules while keeping others enabled. This allows fine-grained control over which instrumentation rules are applied.
```bash
  $ otel set -disable=gorm.json,redis.json
```

Enable All Rules: Enable all rules.
```bash
  $ otel set -disable=
```

Combination of Default and Custom Rules: Use both the default rules and custom rules to provide a comprehensive configuration:
```bash
  $ otel set -rule=custom.json
```

Multiple Rule Files: Combine multiple custom rule files along with the default rules, which can be specified as a comma-separated list:
```bash
  $ otel set -rule=a.json,b.json
```

GoCache Path: Each build will use a temporary directory as go cache, you can set the path to a specific directory to reuse the cache, this
will reduce compilation time by more than 50%.
```bash
  $ otel set -gocache=/tmp/go-cache
```

Using Environment Variables: In addition to using the `otel set` command, configuration can also be overridden using environment variables. For example, the `OTELTOOL_DEBUG` environment variable allows you to force the tool into debug mode temporarily, making this approach effective for one-time configurations without altering permanent settings.

```bash
$ export OTELTOOL_DEBUG=true
$ export OTELTOOL_VERBOSE=true
$ export OTELTOOL_GO_CACHE="/tmp/go-otel-cache"
```

The names of the environment variables correspond to the configuration options available in the `otel set` command with the `OTELTOOL_` prefix.

Full List of Environment Variables:

- `OTELTOOL_DEBUG`: Enable debug mode.
- `OTELTOOL_VERBOSE`: Enable verbose logging.
- `OTELTOOL_RULE_JSON_FILES`: Specify custom rule files.
- `OTELTOOL_DISABLE_RULES`: Disable specific rules. Use 'all' to disable all default rules, or comma-separated list of rule file names to disable specific rules.
- `OTELTOOL_GO_CACHE`: Set the path to a specific directory to reuse the go cache.

This approach provides flexibility for testing changes and experimenting with configurations without permanently altering your existing setup.

## `otel go build`
Once configurations are in place, you can build your project with prefixed `otel` commands. This integrates the tool's configuration directly into the build process:

Standard Build: Build your project with default settings.
```bash
  $ otel go build
```

Output to Specific Location: Build your project and specify an output location.
```bash
  $ otel go build -o app cmd/app
```

Passing Compiler Flags: Use compiler flags for more customized builds.
```bash
  $ otel go build -gcflags="-m" cmd/app
```
No matter how complex your project is, the otel tool simplifies the process by automatically instrumenting your code for effective observability, the only requirement being the addition of the `otel` prefix to your build commands.

## `otel version`

If you want to check the version of the otel tool, you can use the `otel version` command.
```bash
  $ otel version
```