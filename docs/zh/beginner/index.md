# 新手指南：从这里开始

这套指南写给第一次接触 Go 插桩或 OpenTelemetry 的读者。你不需要先了解编译器、AST 或 OpenTelemetry SDK，只需要能读懂简单的 Go 函数。

官方的[用户指南](../index.md)告诉你怎样使用 `otel`，[开发者指南](../dev/overview.md)和[黑客指南](../hacking/overview.md)讲解规则与实现。这里先补上读懂它们所需的基础：每页只回答一个问题，用一个小例子说明，再指向接下来该读的官方页面。

## 学完你能回答

- 程序已经能正常返回，为什么还要给它加“观察代码”？
- 在 `go build` 前加上 `otel`，额外的代码是在什么时候、怎样加进去的？
- 规则（rule）和 hook 分别负责什么？
- 为什么插桩后，构建过程还要下载和准备新的依赖？
- 一次请求怎样变成一条由多个 Span 组成的 Trace？

## 阅读顺序

| 步骤 | 页面 | 要回答的问题 |
| --- | --- | --- |
| 1 | [为什么要插桩](./why-instrument.md) | 有返回值，为什么还不够？ |
| 2 | [包、模块与 go.mod](./modules.md) | Go 怎样找到要编译的代码？ |
| 3 | [构建时与运行时](./build-and-run.md) | `otel` 在哪一步介入？ |
| 4 | [规则与 hook](./rules-and-hooks.md) | 工具怎么知道改哪里、加什么？ |
| 5 | [为什么要补依赖](./dependencies.md) | 多一行调用，为什么要改依赖？ |
| 6 | [Trace 与 Span](./traces-and-spans.md) | 多条记录怎样连成一次请求？ |
| 7 | [动手：第一个自定义 hook](./first-hook.md) | 怎样证明 hook 真的执行了？ |

## 开始前准备

- 使用所选 `otel` 版本支持的 Go 版本；Go 和 OpenTelemetry 的版本要求见[兼容性](../user/compatibility.md)。
- 第 1 到第 6 步只需要阅读，示例可以用普通的 `go run` 运行。
- 第 7 步需要安装 `otel` 工具，安装方法见[概述](../index.md#安装)。

## 示意动画

第 1 到第 6 页嵌有七段自动循环的示意动画，其中第 2 页包含“包和模块”与“go.mod”两段。启用“减少动态效果”或打印时显示最后一步，完整过程也写在正文中。

## 社区资源

社区维护的 [LoongSuite Go 中文学习指南](https://qinghua362330.github.io/loongsuite-go-guide/)提供互动图解和练习。它独立于官方文档维护，是本指南的可选补充。
