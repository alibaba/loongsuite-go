# Plugin migration agent (Loongsuite → OpenTelemetry compile-time instrumentation)

This doc summarizes a repeatable approach for migrating plugins from:

- Loongsuite: `pkg/rules/<plugin>` + `tool/data/rules/<plugin>.json`
- To OpenTelemetry compile-time instrumentation: `pkg/instrumentation/<plugin>`

## Reference: databasesql

- Loongsuite source: `pkg/rules/databasesql` ([repo tree](https://github.com/alibaba/loongsuite-go-agent/tree/main/pkg/rules/databasesql))
- Target location: `pkg/instrumentation/databasesql` ([repo tree](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/tree/main/pkg/instrumentation/databasesql))

## What to compare

### 1) Hook surface (ABI)

In Loongsuite, the “hook ABI” is the combination of:

- Rule manifest entries (`tool/data/rules/<plugin>.json`): function + receiver + `OnEnter`/`OnExit`
- Hook implementations (`pkg/rules/<plugin>/setup.go`): `//go:linkname` functions with the exact signatures used by the injected hook points.

For `databasesql`, the manifest covers:
- Struct field injections (DB/Conn/Tx/Stmt add fields like `Endpoint`, `DriverName`, `DSN`, `Data`)
- Function hooks: `Open`, `PingContext`, `PrepareContext`, `ExecContext`, `QueryContext`, `BeginTx`, plus Conn/Tx/Stmt variants.

### 2) Instrumentation semantics

Verify parity for:
- Span name extraction and kind
- Attributes (semconv): statement, operation, system, server address, collection/table
- Optional metrics listeners

### 3) Parsing/caching behavior

For `databasesql` the source includes:
- DSN parser (`mysql`, `postgres/postgresql`)
- SQL parsing for `collection` and (optional) parameters
- LRU cache to avoid repeated parsing

### 4) Configuration knobs

E.g. env enable/disable flags should keep the same default behavior.

## Migration steps (template)

1. **Inventory** the source plugin (manifest + `setup.go` + instrumenter + helpers).
2. **Port** code into `pkg/instrumentation/<plugin>` and align package paths/imports.
3. **Replace** Loongsuite internal APIs (`pkg/api`, `inst-api*`) with target repo equivalents; keep semconv parity.
4. **Wire** the manifest in the target repo’s format (or generate it if the target uses generators).
5. **Tidy** dependencies and watch for module split conflicts (e.g. genproto monolith vs split).
6. **Add tests**: minimal integration test that triggers the hooks and validates spans/attrs.

## Cursor “agent”

There is a reusable Cursor rule file you can apply when doing these migrations:

- `.cursor/rules/plugin-migration-agent.mdc`

