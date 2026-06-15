# TJump优化

由于trampoline-jump-if和trampoline函数对性能至关重要，我们正在尽力优化它们。trampoline-jump-if的标准形式如下：

```go
if ctx, skip := otel_trampoline_onenter(&arg); skip {
    otel_trampoline_onexit(ctx, &retval)
    return ...
} else {
    defer otel_trampoline_onexit(ctx, &retval)
    ...
}
```

明显的优化机会是在onEnter或onExit钩子不存在的情况下。对于后一种情况，我们可以将defer语句替换为空语句，您可能会说我们可以删除整个else块，但同一函数中可能有多个trampoline-jump-if，它们嵌套在else块中，即：

```go
if ctx, skip := otel_trampoline_onenter(&arg); skip {
    otel_trampoline_onexit(ctx, &retval)
    return ...
} else {
    ;
    ...
}
```

对于前一种情况，情况要复杂一些。我们需要动态地手动构造CallContext，并将其传递给onExit trampoline的defer调用，并将整个条件重写为始终为false。相应的代码片段是：

```go
if false {
    ;
} else {
    defer otel_trampoline_onexit(&CallContext{...}, &retval)
    ...
}
```

if骨架应保持原样，否则trampoline-jump-if的内联将无法工作。在编译期间，dce和sccp过程将删除整个then块。这还不是全部。如果onEnter钩子不使用SkipCall，我们可以进一步优化tjump。在这种情况下，我们可以将trampoline-jump-if的条件重写为始终为false，删除then块中的return语句，它们是内存感知的，并可能在编译期间生成内存SSA值。

```go
if ctx,_ := otel_trampoline_onenter(&arg); false {
    ;
} else {
    defer otel_trampoline_onexit(ctx, &retval)
    ...
}
```

编译器负责将初始化语句提升出if骨架，dce和sccp过程将删除整个then块。所有这些trampoline函数看起来都像是按顺序执行的，即：

```go
ctx,_ := otel_trampoline_onenter(&arg);
defer otel_trampoline_onexit(ctx, &retval)
```

请注意，此优化过程是脆弱的，因为它非常依赖于trampoline-jump-if和trampoline函数的结构。对tjump的任何更改都应仔细检查。

