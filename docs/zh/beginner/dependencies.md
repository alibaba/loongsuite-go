# 5. 为什么要补依赖

**本页问题：** 只是多了一行观察调用，为什么构建时还要准备依赖？

## 手动做一遍

假设项目 `demo` 本来能正常构建：

```go
package main

import "fmt"

func FindRoom(id int) string {
    return "room-42"
}

func main() { fmt.Println(FindRoom(42)) }
```

现在在 `FindRoom` 里加一行调用，使用放在 `demo` 目录外的另一个模块 `example.com/observer`（它的 `Record` 函数只打印一行 `observed: <名称>`）：

```go
package main

import (
	"fmt"
	"example.com/observer"
)

func FindRoom(id int) string {
	observer.Record("FindRoom")
	return "room-42"
}

func main() { fmt.Println(FindRoom(42)) }
```

再次构建会失败：

```text
$ go build .
main.go:5:2: no required module provides package example.com/observer; to add it:
	go get example.com/observer
```

`demo` 的 `go.mod` 没有声明这个依赖，Go 不知道去哪里找 `observer` 的代码。在同一份 `go.mod` 里补上声明后才能构建：

```text
module example.com/demo

go <go-version>

require example.com/observer v0.0.0

replace example.com/observer => ../observer
```

保留 `go mod init` 生成的 `go` 指令；这里的 `<go-version>` 只是占位符，不要原样复制。

`require` 声明需要这个模块；本例的 `replace` 告诉 Go 源码在相邻的 `../observer` 目录。再次构建就能成功。注意 `go build` 只是编译，运行生成的程序时才会执行 `Record`：

```text
$ go build -o demo . && ./demo
observed: FindRoom
room-42
```

原来的返回值不变，只是多了一条观察记录。

![动画：多了一行调用，为什么编译不过](../../public/beginner/deps-zh.svg)

## 回到 otel

`otel` 自动加入的 hook 代码，同样要用到它自己的库，例如 OpenTelemetry SDK 和各个插件的代码。所以工具不仅要把调用加进去，还要让这些代码在构建中能被找到。这就是[预处理阶段](../hacking/preprocess.md)中“准备规则所需依赖”的含义。

> [!NOTE]
> 新加入的代码还可能用到别的库，而这些库又可能命中新的规则，所以准备依赖不一定一次完成。这也是使用 `otel` 后构建变慢的原因之一，见[编译时间](../user/compilation-time.md)。

下一步：[Trace 与 Span](./traces-and-spans.md)
