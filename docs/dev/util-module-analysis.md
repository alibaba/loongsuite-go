# GenAI Util Module Addition Proposal

## 1. Current Project Status Analysis

Loongsuite Go Agent is a zero-code-intrusion OpenTelemetry automatic instrumentation tool with the following characteristics:

- **Zero Code Intrusion**: Achieves transparent instrumentation of applications through compile-time automatic injection, requiring no modification to business code
- **Modular Design**: Supports multiple frameworks and libraries (such as Gin, gRPC, Redis, Kafka, etc.), with each instrumentation rule being an independent Go module
- **GenAI Scenario Support**: Already supports automatic instrumentation for LLM-related libraries including OpenAI, Ollama, LangChain, Eino, MCP, etc.

### Existing Utility Function Distribution

| Directory | Purpose | Suitable as General Runtime Library |
|-----------|---------|-------------------------------------|
| `tool/util/` | Compile-time tool functions (logging, assertions, file operations, etc.) | ❌ Not suitable, tightly coupled with compilation workflow |
| `pkg/inst-api/utils/` | Instrumentation-related utilities (Span operations, attribute extraction, etc.) | ❌ Not suitable, tightly coupled with OpenTelemetry |
| `pkg/inst-api-semconv/` | Semantic convention implementations | ❌ Not suitable, specific to OTel semantic conventions |

**Conclusion**: The project lacks a general-purpose utility function library independent of instrumentation logic. Especially in GenAI/Agent development scenarios, developers need a wide range of general utility functions to handle strings, slices, context, and other operations.

## 2. Recommended Approach: Create Independent Submodule Under pkg/

### Directory Structure

```
pkg/
├── util/                  # New util module
│   ├── go.mod            # Independent go.mod
│   ├── string/           # String utilities
│   │   └── string.go
│   ├── slice/            # Slice utilities
│   │   └── slice.go
│   ├── map/              # Map utilities
│   │   └── map.go
│   ├── time/             # Time utilities
│   │   └── time.go
│   ├── context/          # Context utilities
│   │   └── context.go
│   └── error/            # Error handling utilities
│       └── error.go
```

### Module Definition

```go
module github.com/alibaba/loongsuite-go/pkg/util

go 1.24.0
```

This module exists as an independent submodule with no dependencies on any other internal modules in the project, and can be compiled and published independently.

## 3. Design Principles

1. **Independence**
   - No dependency on project-specific instrumentation logic
   - No dependency on `pkg/api`, `pkg/inst-api`, or other internal modules
   - Minimal external dependencies (rely on the standard library as much as possible)

2. **Generality**
   - Provides commonly used utility functions for LLM Agent development
   - Applicable to any Go project, not limited to the Loongsuite ecosystem
   - Simple and intuitive API design, following idiomatic Go style

3. **Publishability**
   - Can be published as an independent library to Go Module Proxy
   - Other projects can depend on it directly via `go get`
   - Semantic versioning with backward compatibility

4. **High Performance**
   - Zero-allocation or minimal-allocation design
   - Avoids unnecessary reflection
   - Fully leverages Go generics (Go 1.18+)

5. **Type Safety**
   - Uses generics to avoid overuse of `interface{}`
   - Compile-time type checking preferred over runtime checking

## 4. Utility Function Category Planning

| Submodule | Package Name | Main Functions |
|-----------|--------------|----------------|
| `string/` | `utilstr` | String truncation, formatting, template rendering, encoding/decoding, similarity calculation |
| `slice/` | `utilslice` | Slice filtering, mapping, deduplication, pagination, batch operations, grouping |
| `map/` | `utilmap` | Map merging, filtering, key/value extraction, type conversion, deep copy |
| `time/` | `utiltime` | Time formatting, timezone handling, time calculations, human-readable time |
| `context/` | `utilctx` | Context timeout setting, value passing helpers, chaining operations, merging |
| `error/` | `utilerr` | Error wrapping, error classification, error chain tracing, retry determination |

## 5. Usage Examples

### 5.1 string Package Example

```go
package main

import (
	"fmt"

	utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
)

func main() {
	// String truncation (useful for truncating LLM output to prevent excessively long text from consuming too many tokens)
	longOutput := "This is a very long LLM output text that needs to be truncated for display to save space"
	result := utilstr.Truncate(longOutput, 20)
	fmt.Println(result) // "This is a very long ..."

	// Truncation with custom suffix
	result2 := utilstr.TruncateWithSuffix(longOutput, 20, "…[more]")
	fmt.Println(result2)

	// Template rendering (useful for dynamically filling prompt templates)
	prompt := utilstr.Render("Please analyze the following content: {{.Content}}, requirements: {{.Requirement}}", map[string]string{
		"Content":     "user input text",
		"Requirement": "concise and clear",
	})
	fmt.Println(prompt) // "Please analyze the following content: user input text, requirements: concise and clear"

	// Base64 encoding/decoding (useful for multimodal content transmission)
	encoded := utilstr.Base64Encode("hello world")
	fmt.Println(encoded) // "aGVsbG8gd29ybGQ="
	decoded, _ := utilstr.Base64Decode(encoded)
	fmt.Println(decoded) // "hello world"

	// String similarity (useful for intent matching, fuzzy search)
	similarity := utilstr.Similarity("hello", "hallo")
	fmt.Printf("Similarity: %.2f\n", similarity) // 0.80
}
```

### 5.2 slice Package Example

```go
package main

import (
	"fmt"

	utilslice "github.com/alibaba/loongsuite-go/pkg/util/slice"
)

func main() {
	// Slice deduplication (useful for deduplicating retrieval results)
	items := []string{"a", "b", "a", "c", "b"}
	unique := utilslice.Unique(items)
	fmt.Println(unique) // ["a", "b", "c"]

	// Slice filtering (useful for filtering low-confidence results)
	numbers := []int{1, 2, 3, 4, 5, 6}
	even := utilslice.Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Println(even) // [2, 4, 6]

	// Batch processing (useful for LLM batch inference, controlling concurrent request count)
	batches := utilslice.Chunk(numbers, 2)
	fmt.Println(batches) // [[1, 2], [3, 4], [5, 6]]

	// Map transformation (useful for data format conversion)
	doubled := utilslice.Map(numbers, func(n int) int {
		return n * 2
	})
	fmt.Println(doubled) // [2, 4, 6, 8, 10, 12]

	// Reduce aggregation (useful for calculating total token count, etc.)
	sum := utilslice.Reduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Println(sum) // 21

	// Grouping (useful for grouping conversation messages by category)
	type Message struct {
		Role    string
		Content string
	}
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "Help me write code"},
	}
	grouped := utilslice.GroupBy(messages, func(m Message) string {
		return m.Role
	})
	fmt.Println(len(grouped["user"]))      // 2
	fmt.Println(len(grouped["assistant"])) // 1
}
```

### 5.3 map Package Example

```go
package main

import (
	"fmt"

	utilmap "github.com/alibaba/loongsuite-go/pkg/util/map"
)

func main() {
	// Map merging (useful for merging multiple model configurations, overriding default parameters)
	base := map[string]interface{}{
		"model":       "gpt-4",
		"temperature": 0.7,
		"top_p":       0.9,
	}
	override := map[string]interface{}{
		"temperature": 0.9,
		"max_tokens":  1000,
	}
	merged := utilmap.Merge(base, override)
	fmt.Println(merged)
	// {"model": "gpt-4", "temperature": 0.9, "top_p": 0.9, "max_tokens": 1000}

	// Extract all keys (useful for obtaining configuration item lists)
	keys := utilmap.Keys(merged)
	fmt.Println(keys) // ["model", "temperature", "top_p", "max_tokens"]

	// Extract all values
	values := utilmap.Values(merged)
	fmt.Println(values)

	// Filter Map (useful for removing sensitive configuration items)
	filtered := utilmap.Filter(merged, func(k string, v interface{}) bool {
		return k != "temperature"
	})
	fmt.Println(filtered) // {"model": "gpt-4", "top_p": 0.9, "max_tokens": 1000}

	// Map transformation (useful for converting configs to request header format)
	headers := map[string]string{
		"Authorization": "Bearer sk-xxx",
		"Content-Type":  "application/json",
	}
	upperHeaders := utilmap.MapValues(headers, func(k, v string) string {
		if k == "Authorization" {
			return "Bearer [REDACTED]"
		}
		return v
	})
	fmt.Println(upperHeaders)
}
```

### 5.4 time Package Example

```go
package main

import (
	"fmt"
	"time"

	utiltime "github.com/alibaba/loongsuite-go/pkg/util/time"
)

func main() {
	// Format as human-readable time (useful for displaying conversation timestamps)
	t := time.Now().Add(-2 * time.Minute)
	formatted := utiltime.FormatHuman(t)
	fmt.Println(formatted) // "2 minutes ago"

	// Calculate elapsed time (useful for LLM inference timing, performance monitoring)
	start := time.Now()
	// ... perform inference ...
	time.Sleep(100 * time.Millisecond) // Simulate inference latency
	elapsed := utiltime.Since(start)
	fmt.Printf("Inference time: %s\n", elapsed) // "Inference time: 100ms"

	// Detailed elapsed time output
	detail := utiltime.SinceDetail(start)
	fmt.Printf("Elapsed: %dms (%.2fs)\n", detail.Milliseconds, detail.Seconds)

	// Timezone conversion (useful for unifying cross-timezone logs)
	now := time.Now()
	utcTime := utiltime.ToUTC(now)
	localTime := utiltime.ToTimezone(utcTime, "Asia/Shanghai")
	fmt.Println(utcTime.Format(time.RFC3339))
	fmt.Println(localTime.Format(time.RFC3339))

	// Time range check (useful for determining if an API key has expired)
	expireAt := time.Now().Add(24 * time.Hour)
	if utiltime.IsExpired(expireAt) {
		fmt.Println("Expired")
	} else {
		fmt.Println("Not expired")
	}
}
```

### 5.5 context Package Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	utilctx "github.com/alibaba/loongsuite-go/pkg/util/context"
)

func main() {
	// Context with timeout (useful for LLM call timeout control)
	ctx, cancel := utilctx.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Context value passing helper (type-safe key-value storage and retrieval)
	ctx = utilctx.WithValue(ctx, "request_id", "req-12345")
	ctx = utilctx.WithValue(ctx, "user_id", "user-001")
	ctx = utilctx.WithValue(ctx, "model", "gpt-4")

	requestID := utilctx.GetString(ctx, "request_id")
	fmt.Println(requestID) // "req-12345"

	// Batch value setting (useful for initializing request context)
	ctx = utilctx.WithValues(ctx, map[string]interface{}{
		"trace_id":   "trace-abc",
		"session_id": "sess-xyz",
	})

	// Merge cancellation signals from multiple Contexts (useful for multi-path concurrent request scenarios)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel1()
	defer cancel2()

	merged := utilctx.Merge(ctx1, ctx2)
	// When any context is cancelled, merged will also be cancelled
	go func() {
		<-merged.Done()
		fmt.Println("merged context cancelled")
	}()

	cancel1() // Triggers merged cancellation
	time.Sleep(10 * time.Millisecond)
}
```

### 5.6 error Package Example

```go
package main

import (
	"errors"
	"fmt"
	"net"

	utilerr "github.com/alibaba/loongsuite-go/pkg/util/error"
)

func callLLM() error {
	// Simulate LLM call timeout
	return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection timeout")}
}

func main() {
	// Error wrapping (preserves call chain for easy tracing of problem origin)
	err := callLLM()
	if err != nil {
		wrapped := utilerr.Wrap(err, "failed to call LLM")
		fmt.Println(wrapped) // "failed to call LLM: dial tcp: connection timeout"

		// Error wrapping with context information
		detailed := utilerr.Wrapf(err, "failed to call model %s, retry count: %d", "gpt-4", 3)
		fmt.Println(detailed)
	}

	// Error classification (useful for deciding retry strategy)
	if utilerr.IsTimeout(err) {
		fmt.Println("Timeout error, retrying")
	}
	if utilerr.IsRetryable(err) {
		fmt.Println("Retryable error")
	}
	if utilerr.IsNetwork(err) {
		fmt.Println("Network error")
	}

	// Error chain tracing (useful for locating issues in complex call chains)
	wrappedErr := utilerr.Wrap(utilerr.Wrap(err, "layer1"), "layer2")
	chain := utilerr.Chain(wrappedErr)
	for i, e := range chain {
		fmt.Printf("  [%d] %s\n", i, e.Error())
	}
	// [0] layer2: layer1: dial tcp: connection timeout
	// [1] layer1: dial tcp: connection timeout
	// [2] dial tcp: connection timeout

	// Error aggregation (useful for collecting all errors after concurrent requests)
	errs := utilerr.NewMultiError()
	errs.Add(errors.New("model A call failed"))
	errs.Add(errors.New("model B call failed"))
	errs.Add(nil) // nil is ignored
	if errs.HasErrors() {
		fmt.Printf("Total %d errors: %s\n", errs.Len(), errs.Error())
	}
}
```

## 6. Implementation Steps

### Phase 1: Foundation Setup (Week 1)

1. **Create Module Structure**
   - Establish the `pkg/util/` directory and subdirectories
   - Create `go.mod` with declared module path
   - Add `.golangci.yml` code quality configuration

2. **Implement Core Utility Functions**
   - Prioritize implementing `string/`, `slice/`, `error/` — the three most frequently used modules
   - Every function must have complete GoDoc comments
   - Follow Go standard library style

### Phase 2: Feature Completion (Week 2)

3. **Implement Remaining Modules**
   - Implement `map/`, `time/`, `context/` modules
   - Ensure all generic function type constraints are correct

4. **Write Tests**
   - Unit test coverage ≥ 90%
   - Include boundary case tests (empty slices, nil maps, empty strings, etc.)
   - Add benchmark tests to verify performance

### Phase 3: Integration and Release (Week 3)

5. **Update Project Dependencies**
   - Add `require` and `replace` directives in rules modules that need to use it
   - Verify successful compilation

6. **Documentation and Examples**
   - Complete GoDoc documentation
   - Add full examples in the `example/` directory

7. **Independent Release Preparation (Optional)**
   - Configure CI/CD automated testing
   - Semantic version tag management
   - Publish to Go Module Proxy

## 7. Integration Approach

### Internal Project Usage

Reference the util package in rules modules:

```go
// pkg/rules/gin/go.mod
module github.com/alibaba/loongsuite-go/pkg/rules/gin

go 1.24.0

require (
    github.com/alibaba/loongsuite-go/pkg/util v0.0.0-00010101000000-000000000000
)

replace github.com/alibaba/loongsuite-go/pkg/util => ../../util
```

Usage in code:

```go
package gin

import (
    utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
    utilerr "github.com/alibaba/loongsuite-go/pkg/util/error"
)

func processRequest(input string) (string, error) {
    // Use string utility to truncate overly long input
    truncated := utilstr.Truncate(input, 4096)
    
    result, err := doSomething(truncated)
    if err != nil {
        return "", utilerr.Wrap(err, "failed to process request")
    }
    return result, nil
}
```

### External Project Usage

```bash
# Install
go get github.com/alibaba/loongsuite-go/pkg/util@latest

# Use specific sub-packages
go get github.com/alibaba/loongsuite-go/pkg/util/string@latest
```

```go
package main

import (
    utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
    utilslice "github.com/alibaba/loongsuite-go/pkg/util/slice"
)

func main() {
    // Use directly, no initialization required
    result := utilstr.Truncate("hello world", 5)
    unique := utilslice.Unique([]int{1, 2, 2, 3})
    _ = result
    _ = unique
}
```

## 8. Relationship with Existing Modules

```
┌─────────────────────────────────────────────────────────┐
│                    Loongsuite Go Agent                    │
├─────────────────────────────────────────────────────────┤
│  tool/          (compile-time tools, not externally exposed) │
│  ├── instrument/                                         │
│  ├── preprocess/                                         │
│  └── util/       ← compile-time internal tools, unrelated to pkg/util │
├─────────────────────────────────────────────────────────┤
│  pkg/            (runtime libraries)                      │
│  ├── api/        ← OTel API wrapper                      │
│  ├── inst-api/   ← Instrumentation API (depends on OTel) │
│  ├── rules/      ← Framework instrumentation rules       │
│  └── util/       ← [NEW] General utility library (zero external dependencies) │
└─────────────────────────────────────────────────────────┘
```

`pkg/util` sits at the bottom of the dependency chain, depending on no other modules in the project, but can be depended upon by all modules.
