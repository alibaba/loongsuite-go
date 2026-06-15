# 埋点阶段

此阶段使用在预处理期间识别的规则将监控代码注入目标函数。我们将以`net/http`的`(*Transport).RoundTrip()`函数为例，逐步介绍整个过程。

我们的注入机制围绕一个三部分函数模型展开：

1.  **RawFunc**：库中的原始目标函数（例如，`(*Transport).RoundTrip()`）。
2.  **TrampolineFunc**：由我们的工具生成的新函数。它设置监控上下文，处理panic，并调用hook函数。
3.  **HookFunc**：由开发人员提供的实际监控逻辑（例如，`onEnter`/`onExit`探针），其中包含跟踪、指标等逻辑。

总体流程是：**RawFunc → TrampolineFunc → HookFunc**。以下各节详细介绍了我们如何实现这一点。

### 步骤1：将跳转（`tjump`）注入RawFunc

首先，该工具修改`RawFunc`的抽象语法树（AST）。我们在其入口点注入一小段称为**tjump**的跳转代码。

以下是修改后的`(*Transport).RoundTrip()`：

```go
// 原始函数被修改以包含tjump
func (t *Transport) RoundTrip(req *Request) (retVal0 *Response, retVal1 error) {
    // 这是 "tjump"
    if callContext, skip := OtelOnEnterTrampoline_RoundTrip37639(&t, &req); skip {
        // 此块通常为空且很少执行
        return
    } else {
        // 'defer'确保OnExit钩子在函数返回之前运行
        defer OtelOnExitTrampoline_RoundTrip37639(callContext, &retVal0, &retVal1)
    }
    // 'if'语句之后执行原始函数体
    return t.roundTrip(req)
}
```

**关键点：**
*   `if`语句立即调用`TrampolineFunc` (`OtelOnEnter...`)。
*   `else`块总是被执行，因为`skip`几乎总是`false`。这种巧妙的结构使我们能够运行`OnEnter`逻辑，并同时使用`defer`来安排`OnExit`逻辑。
*   `tjump`代码在编译时经过大量优化，以最小化性能开销。（[参见优化详情](./optimize.md)）。

### 步骤2：TrampolineFunc - 准备上下文

`tjump`调用`TrampolineFunc`，它充当一个桥梁。其职责是：
1.  创建一个`CallContext`以在函数之间传递参数和返回值。
2.  设置一个`recover`块以捕获钩子中的任何panic。
3.  调用`HookFunc`（在本例中为`ClientOnEnterImpl`）。

```go
// 此TrampolineFunc由工具生成
func OtelOnEnterTrampoline_RoundTrip37639(t **Transport, req **Request) (*CallContext, bool) {
    // 1. 设置panic恢复
    defer func() {
        if err := recover(); err != nil {
            // 失败钩子的错误处理
        }
    }()

    // 2. 准备上下文
    callContext := &CallContext{
        Params: []interface{}{t, req},
        // ... 其他字段
    }

    // 3. 调用抽象的HookFunc
    ClientOnEnterImpl(callContext, *t, *req)

    return callContext, callContext.SkipCall
}

// 该工具还为HookFunc生成一个无主体的声明。
// 此声明稍后将链接到真实实现。
func ClientOnEnterImpl(callContext *CallContext, t *http.Transport, req *http.Request)
```

### 步骤3：HookFunc - 链接真实的监控逻辑

到目前为止，`ClientOnEnterImpl`只是一个抽象声明。为了将其连接到真实实现，我们使用`go:linkname`指令。这是Go的一个强大功能，允许在编译时按名称链接两个函数。

**开发人员的责任：**

1.  **导入实现**：在一个中心文件（例如，`otel.runtime.go`）中，导入包含hook实现的包。`_`确保包的代码包含在构建中。

    ```go
    package main
    import _ "github.com/your-repo/your-agent/hooks" // 导入hook实现
    ```

2.  **定义和链接hook**：在hooks包中，定义具有实际监控逻辑的函数，并使用`go:linkname`将其连接到工具生成的声明。

    ```go
    package hooks

    //go:linkname clientOnEnter net/http.ClientOnEnterImpl
    func clientOnEnter(call api.CallContext, t *http.Transport, req *http.Request) {
        // 实际的监控代码（跟踪、指标等）放在这里。
        // 例如：开始一个新的span。
    }
    ```
    *注意：该工具会自动将用户友好的名称（`clientOnEnter`）映射到生成的名称（`ClientOnEnterImpl`）。*

### 埋点阶段总结

通过链接这些步骤，我们成功地在不更改原始库源代码的情况下注入了监控代码。整个过程——修改AST、生成trampoline函数和链接hook实现——都由我们的工具在`go build -toolexec`命令期间自动完成。

这种自动化的编译时方法具有显着优势：
*   **非侵入性**：无需手动更改第三方代码。
*   **解耦**：监控逻辑与业务逻辑清晰分离。
*   **健壮**：自动化减少了手动埋点中常见的人为错误。
