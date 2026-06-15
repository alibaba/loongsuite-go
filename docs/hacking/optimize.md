# TJump Optimization

Since trampoline-jump-if and trampoline functions are performance-critical,
we are trying to optimize them as much as possible. The standard form of
trampoline-jump-if looks like

```go
if ctx, skip := otel_trampoline_onenter(&arg); skip {
    otel_trampoline_onexit(ctx, &retval)
    return ...
} else {
    defer otel_trampoline_onexit(ctx, &retval)
    ...
}
```

The obvious optimization opportunities are cases when onEnter or onExit hooks
are not present. For the latter case, we can replace the defer statement to
empty statement, you might argue that we can remove the whole else block, but
there might be more than one trampoline-jump-if in the same function, they are
nested in the else block, i.e.

```go
if ctx, skip := otel_trampoline_onenter(&arg); skip {
    otel_trampoline_onexit(ctx, &retval)
    return ...
} else {
    ;
    ...
}
```

For the former case, it's a bit more complicated. We need to manually construct
CallContext on the fly and pass it to onExit trampoline defer call and rewrite
the whole condition to always false. The corresponding code snippet is

```go
if false {
    ;
} else {
    defer otel_trampoline_onexit(&CallContext{...}, &retval)
    ...
}
```

The if skeleton should be kept as is, otherwise inlining of trampoline-jump-if
will not work. During compiling, the dce and sccp passes will remove the whole
then block. That's not the whole story. We can further optimize the tjump iff
the onEnter hook does not use SkipCall. In this case, we can rewrite condition
of trampoline-jump-if to always false, remove return statement in then block,
they are memory-aware and may generate memory SSA values during compilation.

```go
if ctx,_ := otel_trampoline_onenter(&arg); false {
    ;
} else {
    defer otel_trampoline_onexit(ctx, &retval)
    ...
}
```

The compiler responsible for hoisting the initialization statement out of the
if skeleton, and the dce and sccp passes will remove the whole then block. All
these trampoline functions looks as if they are executed sequentially, i.e.

```go
ctx,_ := otel_trampoline_onenter(&arg);
defer otel_trampoline_onexit(ctx, &retval)
```

Note that this optimization pass is fraigle as it really heavily depends on
the structure of trampoline-jump-if and trampoline functions. Any change in
tjump should be carefully examined.