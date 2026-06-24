# GenAI Util 模块添加方案分析

## 一、项目现状分析

Loongsuite Go Agent 是一个零代码侵入的 OpenTelemetry 自动插桩工具，具备以下特点：

- **零代码侵入**：通过编译期自动注入实现对应用程序的无感知插桩，开发者无需修改业务代码
- **模块化设计**：支持多种框架和库（如 Gin、gRPC、Redis、Kafka 等），每个插桩规则都是独立的 Go 模块
- **GenAI 场景支持**：已支持 OpenAI、Ollama、LangChain、Eino、MCP 等大模型相关库的自动插桩

### 现有工具函数分布

| 目录 | 用途 | 是否适合作为通用运行时库 |
|------|------|------------------------|
| `tool/util/` | 编译时工具函数（日志、断言、文件操作等） | ❌ 不适合，与编译流程紧密耦合 |
| `pkg/inst-api/utils/` | 插桩相关工具（Span 操作、属性提取等） | ❌ 不适合，与 OpenTelemetry 紧密耦合 |
| `pkg/inst-api-semconv/` | 语义约定相关实现 | ❌ 不适合，特定于 OTel 语义约定 |

**结论**：项目中缺少一个通用的、与插桩逻辑无关的工具函数库，尤其在 GenAI/Agent 开发场景中，开发者需要大量通用工具函数来处理字符串、切片、上下文等操作。

## 二、推荐方案：在 pkg/ 下创建独立子模块

### 目录结构

```
pkg/
├── util/                  # 新建的 util 模块
│   ├── go.mod            # 独立的 go.mod
│   ├── string/           # 字符串工具
│   │   └── string.go
│   ├── slice/            # 切片工具
│   │   └── slice.go
│   ├── map/              # Map 工具
│   │   └── map.go
│   ├── time/             # 时间工具
│   │   └── time.go
│   ├── context/          # Context 工具
│   │   └── context.go
│   └── error/            # 错误处理工具
│       └── error.go
```

### 模块定义

```go
module github.com/alibaba/loongsuite-go/pkg/util

go 1.24.0
```

该模块作为独立子模块存在，不依赖项目中的任何其他内部模块，可独立编译和发布。

## 三、设计原则

1. **独立性**
   - 不依赖项目特定的插桩逻辑
   - 不依赖 `pkg/api`、`pkg/inst-api` 等内部模块
   - 最小化外部依赖（尽量只依赖标准库）

2. **通用性**
   - 提供大模型 Agent 开发中常用的工具函数
   - 适用于各种 Go 项目，不限于 Loongsuite 生态
   - API 设计简洁直观，符合 Go 惯用风格

3. **可发布性**
   - 可以作为独立库发布到 Go Module Proxy
   - 其他项目可通过 `go get` 直接依赖
   - 语义化版本管理，向后兼容

4. **高性能**
   - 零分配或最小分配设计
   - 避免不必要的反射
   - 充分利用 Go 泛型（Go 1.18+）

5. **类型安全**
   - 使用泛型避免 `interface{}` 的滥用
   - 编译期类型检查优于运行时检查

## 四、工具函数分类规划

| 子模块 | 包名 | 主要功能 |
|--------|------|---------|
| `string/` | `utilstr` | 字符串截断、格式化、模板渲染、编码/解码、相似度计算 |
| `slice/` | `utilslice` | 切片过滤、映射、去重、分页处理、批量操作、分组 |
| `map/` | `utilmap` | Map 合并、过滤、键值提取、类型转换、深拷贝 |
| `time/` | `utiltime` | 时间格式化、时区处理、时间计算、人类可读时间 |
| `context/` | `utilctx` | Context 超时设置、值传递辅助、链式操作、合并 |
| `error/` | `utilerr` | 错误包装、错误分类、错误链追踪、重试判定 |

## 五、使用示例

### 5.1 string 包示例

```go
package main

import (
	"fmt"

	utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
)

func main() {
	// 字符串截断（适用于大模型输出截断，防止过长文本占用过多 Token）
	longOutput := "这是一段很长的大模型输出文本，需要在展示时进行截断处理以节省空间"
	result := utilstr.Truncate(longOutput, 20)
	fmt.Println(result) // "这是一段很长的大模型输出文本，需要在展..."

	// 带自定义省略符的截断
	result2 := utilstr.TruncateWithSuffix(longOutput, 20, "…[更多]")
	fmt.Println(result2)

	// 模板渲染（适用于 Prompt 模板动态填充）
	prompt := utilstr.Render("请分析以下内容：{{.Content}}，要求：{{.Requirement}}", map[string]string{
		"Content":     "用户输入的文本",
		"Requirement": "简洁明了",
	})
	fmt.Println(prompt) // "请分析以下内容：用户输入的文本，要求：简洁明了"

	// Base64 编码/解码（适用于多模态内容传输）
	encoded := utilstr.Base64Encode("hello world")
	fmt.Println(encoded) // "aGVsbG8gd29ybGQ="
	decoded, _ := utilstr.Base64Decode(encoded)
	fmt.Println(decoded) // "hello world"

	// 字符串相似度（适用于意图匹配、fuzzy search）
	similarity := utilstr.Similarity("hello", "hallo")
	fmt.Printf("相似度: %.2f\n", similarity) // 0.80
}
```

### 5.2 slice 包示例

```go
package main

import (
	"fmt"

	utilslice "github.com/alibaba/loongsuite-go/pkg/util/slice"
)

func main() {
	// 切片去重（适用于去重检索结果）
	items := []string{"a", "b", "a", "c", "b"}
	unique := utilslice.Unique(items)
	fmt.Println(unique) // ["a", "b", "c"]

	// 切片过滤（适用于过滤低置信度结果）
	numbers := []int{1, 2, 3, 4, 5, 6}
	even := utilslice.Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Println(even) // [2, 4, 6]

	// 批量处理（适用于大模型批量推理，控制并发请求数）
	batches := utilslice.Chunk(numbers, 2)
	fmt.Println(batches) // [[1, 2], [3, 4], [5, 6]]

	// Map 映射转换（适用于数据格式转换）
	doubled := utilslice.Map(numbers, func(n int) int {
		return n * 2
	})
	fmt.Println(doubled) // [2, 4, 6, 8, 10, 12]

	// Reduce 聚合（适用于计算总 Token 数等场景）
	sum := utilslice.Reduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Println(sum) // 21

	// 分组（适用于按类别分组对话消息）
	type Message struct {
		Role    string
		Content string
	}
	messages := []Message{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "你好！"},
		{Role: "user", Content: "帮我写代码"},
	}
	grouped := utilslice.GroupBy(messages, func(m Message) string {
		return m.Role
	})
	fmt.Println(len(grouped["user"]))      // 2
	fmt.Println(len(grouped["assistant"])) // 1
}
```

### 5.3 map 包示例

```go
package main

import (
	"fmt"

	utilmap "github.com/alibaba/loongsuite-go/pkg/util/map"
)

func main() {
	// Map 合并（适用于合并多个模型配置、覆盖默认参数）
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

	// 提取所有键（适用于获取配置项列表）
	keys := utilmap.Keys(merged)
	fmt.Println(keys) // ["model", "temperature", "top_p", "max_tokens"]

	// 提取所有值
	values := utilmap.Values(merged)
	fmt.Println(values)

	// 过滤 Map（适用于移除敏感配置项）
	filtered := utilmap.Filter(merged, func(k string, v interface{}) bool {
		return k != "temperature"
	})
	fmt.Println(filtered) // {"model": "gpt-4", "top_p": 0.9, "max_tokens": 1000}

	// Map 转换（适用于将配置转为请求头格式）
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

### 5.4 time 包示例

```go
package main

import (
	"fmt"
	"time"

	utiltime "github.com/alibaba/loongsuite-go/pkg/util/time"
)

func main() {
	// 格式化为人类可读时间（适用于对话时间展示）
	t := time.Now().Add(-2 * time.Minute)
	formatted := utiltime.FormatHuman(t)
	fmt.Println(formatted) // "2分钟前"

	// 计算耗时（适用于大模型推理计时、性能监控）
	start := time.Now()
	// ... 执行推理 ...
	time.Sleep(100 * time.Millisecond) // 模拟推理耗时
	elapsed := utiltime.Since(start)
	fmt.Printf("推理耗时: %s\n", elapsed) // "推理耗时: 100ms"

	// 带格式的耗时输出
	detail := utiltime.SinceDetail(start)
	fmt.Printf("耗时: %dms (%.2fs)\n", detail.Milliseconds, detail.Seconds)

	// 时区转换（适用于跨时区日志统一）
	now := time.Now()
	utcTime := utiltime.ToUTC(now)
	localTime := utiltime.ToTimezone(utcTime, "Asia/Shanghai")
	fmt.Println(utcTime.Format(time.RFC3339))
	fmt.Println(localTime.Format(time.RFC3339))

	// 时间范围判定（适用于判断 API Key 是否过期）
	expireAt := time.Now().Add(24 * time.Hour)
	if utiltime.IsExpired(expireAt) {
		fmt.Println("已过期")
	} else {
		fmt.Println("未过期")
	}
}
```

### 5.5 context 包示例

```go
package main

import (
	"context"
	"fmt"
	"time"

	utilctx "github.com/alibaba/loongsuite-go/pkg/util/context"
)

func main() {
	// 带超时的 Context（适用于大模型调用超时控制）
	ctx, cancel := utilctx.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Context 值传递辅助（类型安全的键值存取）
	ctx = utilctx.WithValue(ctx, "request_id", "req-12345")
	ctx = utilctx.WithValue(ctx, "user_id", "user-001")
	ctx = utilctx.WithValue(ctx, "model", "gpt-4")

	requestID := utilctx.GetString(ctx, "request_id")
	fmt.Println(requestID) // "req-12345"

	// 批量设置值（适用于初始化请求上下文）
	ctx = utilctx.WithValues(ctx, map[string]interface{}{
		"trace_id":   "trace-abc",
		"session_id": "sess-xyz",
	})

	// 合并多个 Context 的取消信号（适用于多路并发请求场景）
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel1()
	defer cancel2()

	merged := utilctx.Merge(ctx1, ctx2)
	// 任一 context 取消，merged 都会取消
	go func() {
		<-merged.Done()
		fmt.Println("merged context cancelled")
	}()

	cancel1() // 触发 merged 取消
	time.Sleep(10 * time.Millisecond)
}
```

### 5.6 error 包示例

```go
package main

import (
	"errors"
	"fmt"
	"net"

	utilerr "github.com/alibaba/loongsuite-go/pkg/util/error"
)

func callLLM() error {
	// 模拟大模型调用超时
	return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection timeout")}
}

func main() {
	// 错误包装（保留调用链，方便追踪问题来源）
	err := callLLM()
	if err != nil {
		wrapped := utilerr.Wrap(err, "调用大模型失败")
		fmt.Println(wrapped) // "调用大模型失败: dial tcp: connection timeout"

		// 带上下文信息的错误包装
		detailed := utilerr.Wrapf(err, "调用模型 %s 失败，重试次数: %d", "gpt-4", 3)
		fmt.Println(detailed)
	}

	// 错误分类（适用于决定重试策略）
	if utilerr.IsTimeout(err) {
		fmt.Println("超时错误，进行重试")
	}
	if utilerr.IsRetryable(err) {
		fmt.Println("可重试错误")
	}
	if utilerr.IsNetwork(err) {
		fmt.Println("网络错误")
	}

	// 错误链追踪（适用于复杂调用链路的问题定位）
	wrappedErr := utilerr.Wrap(utilerr.Wrap(err, "layer1"), "layer2")
	chain := utilerr.Chain(wrappedErr)
	for i, e := range chain {
		fmt.Printf("  [%d] %s\n", i, e.Error())
	}
	// [0] layer2: layer1: dial tcp: connection timeout
	// [1] layer1: dial tcp: connection timeout
	// [2] dial tcp: connection timeout

	// 错误聚合（适用于并发请求后收集所有错误）
	errs := utilerr.NewMultiError()
	errs.Add(errors.New("模型 A 调用失败"))
	errs.Add(errors.New("模型 B 调用失败"))
	errs.Add(nil) // nil 会被忽略
	if errs.HasErrors() {
		fmt.Printf("共 %d 个错误: %s\n", errs.Len(), errs.Error())
	}
}
```

## 六、实施步骤

### 阶段一：基础搭建（第 1 周）

1. **创建模块结构**
   - 建立 `pkg/util/` 目录及子目录
   - 创建 `go.mod`，声明模块路径
   - 添加 `.golangci.yml` 代码质量配置

2. **实现核心工具函数**
   - 优先实现 `string/`、`slice/`、`error/` 三个最高频模块
   - 每个函数必须有完整的 GoDoc 注释
   - 遵循 Go 标准库风格

### 阶段二：完善功能（第 2 周）

3. **补充剩余模块**
   - 实现 `map/`、`time/`、`context/` 模块
   - 确保所有泛型函数类型约束正确

4. **编写测试**
   - 单元测试覆盖率 ≥ 90%
   - 包含边界情况测试（空切片、nil map、空字符串等）
   - 添加基准测试（Benchmark）验证性能

### 阶段三：集成与发布（第 3 周）

5. **更新项目依赖**
   - 在需要使用的 rules 模块中添加 `require` 和 `replace` 指令
   - 验证编译通过

6. **文档和示例**
   - 完善 GoDoc 文档
   - 在 `example/` 目录下添加完整示例

7. **独立发布准备（可选）**
   - 配置 CI/CD 自动化测试
   - 语义化版本标签管理
   - 发布到 Go Module Proxy

## 七、集成方式

### 项目内部使用

在 rules 模块中引用 util 包：

```go
// pkg/rules/gin/go.mod
module github.com/alibaba/loongsuite-go/pkg/rules/gin

go 1.24.0

require (
    github.com/alibaba/loongsuite-go/pkg/util v0.0.0-00010101000000-000000000000
)

replace github.com/alibaba/loongsuite-go/pkg/util => ../../util
```

在代码中使用：

```go
package gin

import (
    utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
    utilerr "github.com/alibaba/loongsuite-go/pkg/util/error"
)

func processRequest(input string) (string, error) {
    // 使用字符串工具截断过长输入
    truncated := utilstr.Truncate(input, 4096)
    
    result, err := doSomething(truncated)
    if err != nil {
        return "", utilerr.Wrap(err, "处理请求失败")
    }
    return result, nil
}
```

### 外部项目使用

```bash
# 安装
go get github.com/alibaba/loongsuite-go/pkg/util@latest

# 使用特定子包
go get github.com/alibaba/loongsuite-go/pkg/util/string@latest
```

```go
package main

import (
    utilstr "github.com/alibaba/loongsuite-go/pkg/util/string"
    utilslice "github.com/alibaba/loongsuite-go/pkg/util/slice"
)

func main() {
    // 直接使用，无需任何初始化
    result := utilstr.Truncate("hello world", 5)
    unique := utilslice.Unique([]int{1, 2, 2, 3})
    _ = result
    _ = unique
}
```

## 八、与现有模块的关系

```
┌─────────────────────────────────────────────────────────┐
│                    Loongsuite Go Agent                    │
├─────────────────────────────────────────────────────────┤
│  tool/          (编译时工具，不对外暴露)                    │
│  ├── instrument/                                         │
│  ├── preprocess/                                         │
│  └── util/       ← 编译时内部工具，与 pkg/util 无关       │
├─────────────────────────────────────────────────────────┤
│  pkg/            (运行时库)                               │
│  ├── api/        ← OTel API 封装                         │
│  ├── inst-api/   ← 插桩 API (依赖 OTel)                  │
│  ├── rules/      ← 各框架插桩规则                         │
│  └── util/       ← 【新增】通用工具库 (零外部依赖)        │
└─────────────────────────────────────────────────────────┘
```

`pkg/util` 处于依赖链的最底层，不依赖项目中的任何其他模块，但可以被所有模块依赖。
