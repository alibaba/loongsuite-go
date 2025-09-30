# Preprocess Phase

In this phase, the tool analyzes the third-party dependencies in the user's project. It matches them against a repository of instrumentation rules to identify which ones are applicable. Concurrently, it prepares any additional dependencies that these rules require.

Instrumentation rules precisely define what code to inject, where to inject it, and for which version of a framework or standard library. Different types of rules serve different purposes, including:

- InstFuncRule: Inject code at the entry and exit points of a method.
- InstStructRule: Modify a struct by adding a new field.
- InstFileRule: Add a new file to participate in the original compilation process.

Once preprocessing is complete, the tool initiates the instrumented build by executing `go build -toolexec otel cmd/app`. The `-toolexec` flag is central to our automatic instrumentation. It instructs the Go compiler to replace its standard build tools with our custom tool, **otel**.

This handoff to the otel tool marks the transition to the [Instrument Phase](./instrument.md), where the rules identified during preprocessing will be applied.