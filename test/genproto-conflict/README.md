# Genproto Conflict Test

This test case validates the fix for [issue #627](https://github.com/alibaba/loongsuite-go-agent/issues/627).

## Problem

When building projects that use both:
1. Libraries with old monolithic `google.golang.org/genproto` dependency (e.g., `aliyun-odps-go-sdk`)
2. OpenTelemetry libraries that use the newer split modules:
   - `google.golang.org/genproto/googleapis/api`
   - `google.golang.org/genproto/googleapis/rpc`

The Go build would fail with "ambiguous import" errors like:

```
ambiguous import: found package google.golang.org/genproto/googleapis/rpc/errdetails in multiple modules:
  google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1
  google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a
```

## Solution

The fix adds explicit version constraints for the split `google.golang.org/genproto` modules in `tool/preprocess/update.go`:

```go
"google.golang.org/genproto/googleapis/api": "v0.0.0-20251202230838-ff82c1b0f217",
"google.golang.org/genproto/googleapis/rpc": "v0.0.0-20251202230838-ff82c1b0f217",
```

When the otel tool runs `go mod tidy`, it now forces all dependencies to use the newer split module versions, eliminating the ambiguity.

## Test Content

This test:
1. Uses `go.uber.org/zap` for logging (a common dependency)
2. Imports `github.com/aliyun/aliyun-odps-go-sdk/odps` which brings in old genproto
3. Verifies the build succeeds without ambiguous import errors
4. Checks that the debug log doesn't contain error messages about conflicting genproto versions

## Running the Test

```bash
cd test
go test -v -run TestGenprotoConflict
```

Or run with a specific plugin name:

```bash
TEST_PLUGIN_NAME=genproto-conflict go test -v
```
