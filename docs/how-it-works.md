# How it works

Under normal circumstances, the `go build` command goes through the following main steps to compile a Golang application:

1. Source Code Parsing: The Golang compiler first parses the source code files and transforms them into an Abstract Syntax Tree (AST).
2. Type Checking: After parsing, type checking ensures that the code adheres to Golang's type system.
3. Semantic Analysis: This involves analyzing the semantics of the program, including variable definitions and usages, as well as package imports.
4. Compilation Optimization: The syntax tree is converted into an intermediate representation and various optimizations are performed to improve code execution efficiency.
5. Code Generation: Machine code for the target platform is generated.
6. Linking: Different packages and libraries are linked together to form a single executable file.

When using our automatic instrumentation tool, two additional phases are added before the above steps: **Preprocessing** and **Instrument**.

![](images/workflow.png)

- `Preprocess`: Analyze dependencies and select rules that should be used later.
- `Instrument`: Generate code based on rules and inject new code into source code.

## Preprocess
In this phase, the tool analyzes third-party library dependencies in the user's project 
code and matches them against existing instrumentation rules to find appropriate rules. 
It also pre-configures the extra dependencies required by these rules.

Instrumentation rules precisely define which code needs to be injected into which 
version of which framework or standard library. Different types of instrumentation 
rules serve different purposes. The currently available types of instrumentation
rules include:

- InstFuncRule: Inject code at the entry and exit points of a method.
- InstStructRule: Modify a struct by adding a new field.
- InstFileRule: Add a new file to participate in the original compilation process.

Once all the preprocessing is complete, `go build -toolexec otel cmd/app` 
is called for compilation. The `-toolexec` parameter is the core of our automatic 
instrumentation, used to intercept the conventional build process and replace it
with a user-defined tool, allowing developers to customize the build process more 
flexibly. Here, `otel` is the automatic instrumentation tool,
which brings us to the Instrument phase.

## Instrument
During this phase, trampoline code is inserted into target functions based on the rules. 
Trampoline code is essentially a complex *If-statement* that allows the insertion of 
monitoring code at the entry and exit points of the target function, enabling the 
collection of monitoring data. Additionally, several optimizations are performed at
the AST level to minimize the extra performance overhead of the trampoline code and 
optimize code execution efficiency.

After these steps are completed, the tool modifies the compilation parameters and then
calls `go build cmd/app` for normal compilation, as described earlier.

# `net/http` example
First, we classify the following three types of functions: *RawFunc*, *TrampolineFunc*, *HookFunc*. RawFunc is the original function that needs to be injected. TrampolineFunc is the trampoline function. HookFunc is onEnter/onExit functions that need to be inserted at the entry and exit points of the original function as probe code. RawFunc jumps to TrampolineFunc via the inserted trampoline code, then TrampolineFunc constructs the context, prepares the error recovery handling, and finally jumps to HookFunc to execute the probe code.

![](images/tjump.png)

Next, we use `net/http` as an example to demonstrate how compile-time automatic instrumentation can insert monitoring code into the target function `(*Transport).RoundTrip()`. The framework will generate trampoline code at the entry of this function, which is an if statement (actually one line, written in multiple lines for demonstration) that jumps to TrampolineFunc:

```go
func (t *Transport) RoundTrip(req *Request) (retVal0 *Response, retVal1 error) {
    if callContext37639, _ := OtelOnEnterTrampoline_RoundTrip37639(&t, &req); false {
    } else {
        defer OtelOnExitTrampoline_RoundTrip37639(callContext37639, &retVal0, &retVal1)
    }
    return t.roundTrip(req)
}
```

Here, `OtelOnEnterTrampoline_RoundTrip37639` is the TrampolineFunc. It prepares error handling and the call context, then jumps to `ClientOnEnterImpl`:

```go
func OtelOnEnterTrampoline_RoundTrip37639(t **Transport, req **Request) (*CallContext, bool) {
    defer func() {
        if err := recover(); err != nil {
            println("failed to exec onEnter hook", "clientOnEnter")
            if e, ok := err.(error); ok {
                println(e.Error())
            }
            fetchStack, printStack := OtelGetStackImpl, OtelPrintStackImpl
            if fetchStack != nil && printStack != nil {
                printStack(fetchStack())
            }
        }
    }()
    callContext := &CallContext{
        Params:     nil,
        ReturnVals: nil,
        SkipCall:   false,
    }
    callContext.Params = []interface{}{t, req}
    ClientOnEnterImpl(callContext, *t, *req)
    return callContext, callContext.SkipCall
}
```

The `ClientOnEnterImpl` is the HookFunc, which is our probe code where traces, metrics reporting, etc., are performed. `ClientOnEnterImpl` is a function pointer, pre-configured in the automatically generated *otel_setup_inst.go* during the preprocessing stage, and it actually points to `clientOnEnter`:

```go
// == otel_setup_inst.go
package otel_rules

import http328 "net/http"
...

func init() {
    http328.ClientOnEnterImpl = clientOnEnter
    ...
}
```

The `clientOnEnter` function performs the actual monitoring tasks:

```go
// == otel_rule_http59729.go
func clientOnEnter(call api.CallContext, t *http.Transport, req *http.Request) {
    ...
    var tracer trace.Tracer
    if span := trace.SpanFromContext(req.Context()); span.SpanContext().IsValid() {
        tracer = span.TracerProvider().Tracer("")
    } else {
        tracer = otel.GetTracerProvider().Tracer("")
    }
    opts := append([]trace.SpanStartOption{}, trace.WithSpanKind(trace.SpanKindClient))
    ctx, span := tracer.Start(req.Context(), req.URL.Path, opts...)
    var attrs []attribute.KeyValue
    attrs = append(attrs, semconv.HTTPMethodKey.String(req.Method))
    attrs = append(attrs, attributes.MakeSpanAttrs(req.URL.Path, req.URL.Host, attributes.Http)...)
    span.SetAttributes(attrs...)
    bag := baggage.FromContext(ctx)
    if mem, err := baggage.NewMemberRaw(constants.BAGGAGE_PARENT_PID, attributes.Pid); err == nil {
        bag, _ = bag.SetMember(mem)
    }
    if mem, err := baggage.NewMemberRaw(constants.BAGGAGE_PARENT_RPC, sdktrace.GetRpc()); err == nil {
        bag, _ = bag.SetMember(mem)
    }
    sdktrace.SetGLocalData(constants.TRACE_ID, span.SpanContext().TraceID().String())
    sdktrace.SetGLocalData(constants.SPAN_ID, span.SpanContext().SpanID().String())
    ctx = baggage.ContextWithBaggage(ctx, bag)
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    req = req.WithContext(ctx)
    *(call.Params[1].(**http.Request)) = req
    return
}
```

Through the above steps, we not only inserted monitoring code into the `(*Transport).RoundTrip()` function but also ensured the accuracy and propagation of monitoring data and context. During compile-time automatic instrumentation, these operations are all done automatically, saving developers a significant amount of time and reducing the error rate of manual probes.
