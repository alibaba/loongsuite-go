# franz-go 插件开源迁移总结

## 背景

将内部 ARMS 仓库中的 `franz-go`（`github.com/twmb/franz-go`）Kafka 客户端自动插桩插件迁移至开源社区仓库 [loongsuite-go-agent](https://github.com/alibaba/loongsuite-go-agent)，对应 Issue: https://github.com/alibaba/loongsuite-go-agent/issues/570

## 迁移范围

### 内部版本 vs 开源版本

| 维度 | 内部版本 (ARMS) | 开源版本 (loongsuite) |
|------|-----------------|----------------------|
| OTel SDK | 内部封装 `aotel` | 标准 `go.opentelemetry.io/otel` |
| 消息语义 | 内部 `arms-semconv` | 标准 `inst-api-semconv/instrumenter/message` |
| 监控指标 | ARMS MetricListener | 无（纯 Tracing） |
| 配置管理 | `arms_config` | 环境变量 `OTEL_FRANZ_GO_ENABLED` |
| Go 版本 | go 1.22 | go 1.24 |

### 核心改动

1. **移除所有 ARMS 内部依赖**：`aotel`、`arms_config`、`arms-semconv` 等
2. **使用标准 OTel API**：`go.opentelemetry.io/otel`、`go.opentelemetry.io/otel/sdk`
3. **使用开源 instrumenter 框架**：`pkg/inst-api/instrumenter`、`pkg/inst-api-semconv/instrumenter/message`
4. **重新实现属性提取器**：实现 `message.MessageAttrsGetter` 接口
5. **适配测试框架**：使用 `test/verifier` + `tracetest.SpanStubs` 验证 Trace

## 文件清单

### 规则定义

```
tool/data/rules/franz.json
```

拦截点：`kgo.NewClient`，通过 Hook 机制注入 Tracer，无需拦截每次 produce/consume 调用。

### 插件代码

```
pkg/rules/franz-go/
├── go.mod                       # 模块定义，依赖 franz-go v1.18.0
├── go.sum
├── kafka_data_type.go           # Producer/Consumer 请求响应模型
├── carrier.go                   # RecordCarrier - 通过 Record Header 传播上下文
├── kafka_otel_instrumenter.go   # OTel Instrumenter 构建 + 属性/状态提取器
├── kafka_tracer.go              # kgo.Hook 接口实现（4个 Hook）
└── setup.go                     # go:linkname 入口，拦截 NewClient
```

### Scope 常量

```
pkg/inst-api/utils/scope.go  (新增 2 行)
```

```go
const FRANZ_GO_PRODUCER_SCOPE_NAME = "loongsuite.instrumentation.franz-go"
const FRANZ_GO_CONSUMER_SCOPE_NAME = "loongsuite.instrumentation.franz-go"
```

### 测试代码

```
test/franz_go_tests.go                    # 测试入口 + testcontainers Kafka 初始化
test/franzkafka/v1.18.0/
├── go.mod
├── go.sum
└── test_franz.go                         # 集成测试：produce + consume + trace 验证
```

## 插桩设计

### 架构

```
kgo.NewClient()
    ↓ go:linkname hook
franzNewClientOnEnter()
    ↓ 注入 kgo.WithHooks(KafkaTracer{})
    
KafkaTracer 实现 4 个 Hook 接口：
├── HookProduceRecordBuffered   → Start producer span + inject context to headers
├── HookProduceRecordUnbuffered → End producer span
├── HookFetchRecordBuffered     → Start consumer span + extract context from headers
└── HookFetchRecordUnbuffered   → End consumer span
```

### 上下文传播

- **Producer**：通过 `RecordCarrier.Set()` 将 trace context 写入 `kgo.Record.Headers`
- **Consumer**：通过 `RecordCarrier.Get()` 从 `kgo.Record.Headers` 读取 trace context
- 实现 `propagation.TextMapCarrier` 接口，兼容 W3C TraceContext / B3 等标准传播格式

### Span 属性

Producer Span:
- `messaging.system` = `kafka`
- `messaging.destination.name` = topic name
- `messaging.operation.name` = `publish`
- SpanKind = PRODUCER

Consumer Span:
- `messaging.system` = `kafka`
- `messaging.destination.name` = topic name
- `messaging.operation.name` = `process`
- `messaging.message.body.size` = record value length
- SpanKind = CONSUMER

### Trace 验证

测试验证 producer span 和 consumer span 属于同一个 trace（通过 Header 传播 traceId），确保：
- span name 格式为 `{topic} {operation}`
- 属性符合 OTel Messaging 语义规范
- producer → consumer 的 parent-child 关系正确

## 测试环境

- 使用 `testcontainers-go` 启动 **KRaft 模式** Kafka（`apache/kafka:3.7.0`，无需 Zookeeper）
- 端口固定绑定 `9092`，通过环境变量 `FRANZ_KAFKA_PORT` 传递给测试应用

## 兼容性

- franz-go 版本：`[1.18.0,)` — 支持 v1.18.0 及以上所有版本
- Go 版本：1.24+
- 通过环境变量 `OTEL_FRANZ_GO_ENABLED=false` 可禁用插桩
