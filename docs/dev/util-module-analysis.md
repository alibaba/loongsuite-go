# GenAI Util 模块添加方案分析

## 项目调研总结

### 1. 项目概述

**Loongsuite Go Agent** 是一个 Go 语言的 OpenTelemetry 自动插桩工具，主要特点：

- **零代码侵入**：在编译时自动为 Go 应用添加可观测性支持
- **模块化设计**：支持多种框架和库的自动插桩（如 Gin、GORM、Redis、gRPC 等）
- **独立模块架构**：每个插桩规则都是独立的 Go 模块

### 2. 项目结构分析

```
loongsuite-go-agent/
├── pkg/                    # 核心包目录
│   ├── api/               # API 定义（CallContext 接口等）
│   ├── inst-api/          # 插桩 API
│   │   ├── instrumenter/  # 插桩器实现
│   │   └── utils/         # 插桩相关工具函数
│   ├── inst-api-semconv/  # 语义约定
│   ├── rules/             # 各种库的插桩规则（每个都是独立模块）
│   │   ├── gin/
│   │   ├── gorm/
│   │   ├── redis/
│   │   └── ...
│   └── go.mod             # pkg 模块的 go.mod
├── tool/                  # 编译时工具
│   └── util/              # 工具函数（编译时使用，不适合作为库）
├── example/               # 示例代码
└── test/                  # 测试代码
```

### 3. 模块组织方式

#### 3.1 模块路径规范

- **主模块**：`github.com/alibaba/loongsuite-go-agent`
- **核心包模块**：`github.com/alibaba/loongsuite-go-agent/pkg`
- **规则模块**：`github.com/alibaba/loongsuite-go-agent/pkg/rules/{name}`

#### 3.2 模块依赖关系

每个规则模块（如 `pkg/rules/gin`）都有自己的 `go.mod`，通过 `replace` 指令引用主 pkg 模块：

```go
module github.com/alibaba/loongsuite-go-agent/pkg/rules/gin

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg

require (
    github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
    github.com/gin-gonic/gin v1.10.0
    // ...
)
```

### 4. 现有工具函数分析

#### 4.1 `tool/util/` - 编译时工具函数

**位置**：`tool/util/`

**特点**：
- 用于编译时工具（`otel` 命令）
- 包含文件操作、命令执行、路径处理等
- **不适合作为运行时库**，因为依赖编译时上下文

**主要功能**：
- 文件/目录操作（CopyFile, CopyDir, ReadFile, WriteFile）
- 命令执行（RunCmd）
- 路径判断（IsGoFile, IsGoModFile 等）
- 编译命令解析（SplitCompileCmds, FindFlagValue）

#### 4.2 `pkg/inst-api/utils/` - 插桩相关工具

**位置**：`pkg/inst-api/utils/`

**特点**：
- 属于插桩 API 的一部分
- 包含过滤器、作用域名称等插桩特定功能
- 与 OpenTelemetry 插桩逻辑紧密耦合

**主要功能**：
- URL 过滤器（UrlFilter, SpanNameFilter）
- 作用域名称常量（各种库的 SCOPE_NAME）
- 插桩元数据管理

## Util 模块设计方案

### 方案一：在 `pkg/` 下创建独立子模块（推荐）

#### 1.1 目录结构

```
pkg/
├── util/                  # 新建的 util 模块
│   ├── go.mod            # 独立的 go.mod
│   ├── go.sum
│   ├── string/           # 字符串工具
│   ├── slice/            # 切片工具
│   ├── map/              # Map 工具
│   ├── time/             # 时间工具
│   ├── context/          # Context 工具
│   ├── error/             # 错误处理工具
│   └── ...               # 其他通用工具
├── api/
├── inst-api/
└── rules/
```

#### 1.2 模块定义

**文件**：`pkg/util/go.mod`

```go
module github.com/alibaba/loongsuite-go-agent/pkg/util

go 1.24.0

require (
    // 最小化依赖，只包含必要的标准库和第三方库
    // 避免引入 OpenTelemetry 等重型依赖
)
```

#### 1.3 设计原则

1. **独立性**：
   - 不依赖项目特定的插桩逻辑
   - 不依赖 `pkg/api`、`pkg/inst-api` 等内部模块
   - 最小化外部依赖

2. **通用性**：
   - 提供大模型 agent 开发中常用的工具函数
   - 适用于各种 Go 项目，不仅限于 OpenTelemetry 插桩

3. **可发布性**：
   - 可以作为独立库发布到 GitHub/GitLab
   - 其他项目可以通过 `go get` 直接依赖

#### 1.4 示例工具函数分类

**字符串工具** (`pkg/util/string/`)：
- 字符串截断、格式化
- 模板渲染辅助
- 编码/解码工具

**切片工具** (`pkg/util/slice/`)：
- 切片过滤、映射、去重
- 分页处理
- 批量操作

**Map 工具** (`pkg/util/map/`)：
- Map 合并、过滤
- 键值提取
- 类型转换

**时间工具** (`pkg/util/time/`)：
- 时间格式化
- 时区处理
- 时间计算

**Context 工具** (`pkg/util/context/`)：
- Context 超时设置
- Context 值传递辅助
- Context 链式操作

**错误处理** (`pkg/util/error/`)：
- 错误包装
- 错误分类
- 错误链追踪

### 方案二：作为 `pkg/` 的子包（不推荐）

如果 util 模块需要依赖 `pkg/api` 或其他内部模块，可以作为 `pkg/util` 子包，但这样会：
- 增加模块间的耦合
- 限制独立发布的可能性
- 其他项目需要依赖整个 `pkg` 模块

## 实施步骤

### 步骤 1：创建模块结构

```bash
mkdir -p pkg/util/{string,slice,map,time,context,error}
cd pkg/util
go mod init github.com/alibaba/loongsuite-go-agent/pkg/util
```

### 步骤 2：添加初始工具函数

根据实际需求，添加常用的工具函数。建议从最通用的开始：

1. **字符串工具**：最常用，依赖最少
2. **切片工具**：数据处理常用
3. **错误处理**：Go 项目必备

### 步骤 3：编写测试

为每个工具函数编写单元测试，确保：
- 功能正确性
- 边界情况处理
- 性能考虑

### 步骤 4：更新项目依赖

如果项目内部需要使用 util 模块，在相关模块的 `go.mod` 中添加：

```go
require (
    github.com/alibaba/loongsuite-go-agent/pkg/util v0.0.0-00010101000000-000000000000
)

replace github.com/alibaba/loongsuite-go-agent/pkg/util => ../../util
```

### 步骤 5：文档和示例

1. 创建 `pkg/util/README.md`，说明：
   - 模块用途
   - 功能列表
   - 使用示例
   - API 文档

2. 在 `example/` 下创建 util 使用示例

### 步骤 6：独立发布准备（可选）

如果计划独立发布：

1. **创建独立仓库**（或使用 monorepo 的 submodule）
2. **版本管理**：使用语义化版本（v1.0.0, v1.1.0 等）
3. **CI/CD**：设置自动发布流程
4. **文档站点**：使用 pkg.go.dev 或自建文档

## 与其他模块的集成

### 在规则模块中使用

规则模块可以通过以下方式使用 util：

```go
// pkg/rules/gin/go.mod
require (
    github.com/alibaba/loongsuite-go-agent/pkg/util v0.0.0-00010101000000-000000000000
)

replace github.com/alibaba/loongsuite-go-agent/pkg/util => ../../../util
```

```go
// pkg/rules/gin/setup.go
import (
    "github.com/alibaba/loongsuite-go-agent/pkg/util/string"
    "github.com/alibaba/loongsuite-go-agent/pkg/util/error"
)
```

### 外部项目使用

外部项目可以直接依赖：

```bash
go get github.com/alibaba/loongsuite-go-agent/pkg/util@latest
```

```go
import (
    "github.com/alibaba/loongsuite-go-agent/pkg/util/string"
    "github.com/alibaba/loongsuite-go-agent/pkg/util/slice"
)
```

## 注意事项

### 1. 依赖管理

- **最小化依赖**：只引入必要的标准库和轻量级第三方库
- **避免循环依赖**：util 模块不应依赖 `pkg/api`、`pkg/inst-api` 等
- **版本兼容性**：保持 Go 版本要求与主项目一致（当前为 1.24.0）

### 2. 命名规范

- 遵循 Go 命名规范
- 包名简洁明了（如 `stringutil`、`sliceutil`）
- 函数名清晰表达功能

### 3. 性能考虑

- 避免不必要的内存分配
- 对于频繁调用的函数，考虑性能优化
- 提供性能测试基准

### 4. 向后兼容

- 一旦发布，保持 API 的向后兼容性
- 重大变更需要版本升级（v1 -> v2）

## 总结

推荐采用**方案一**：在 `pkg/` 下创建独立的 `util` 子模块。

**优势**：
- ✅ 模块化设计，符合项目现有架构
- ✅ 可以独立发布，供其他项目使用
- ✅ 最小化依赖，保持轻量级
- ✅ 易于维护和扩展

**实施建议**：
1. 先实现最常用的工具函数（字符串、切片、错误处理）
2. 逐步扩展功能，根据实际需求添加
3. 保持 API 简洁，避免过度设计
4. 充分测试，确保质量
