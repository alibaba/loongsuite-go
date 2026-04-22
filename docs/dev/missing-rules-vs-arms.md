# Missing Rules Compared to ARMS Repo

- Source repo compared: `/Users/zhb192729/arms/otel-go-auto-instrumentation/tool/data/rules`
- Current repo: `/Users/zhb192729/github/loongsuite-go-agent/tool/data/rules`
- Generated at: 2026-04-22 14:43:28
- Missing rule files count: **37**
- Possibly covered by current rules (same `ImportPath` overlap): **11**
- Fully missing (no `ImportPath` overlap): **26**

## Classification

### 1) Possibly Covered (Same ImportPath, Different Rule File Name)

| Missing Rule File | Overlapping ImportPath(s) | Existing Rule File(s) |
|---|---|---|
| `clickhouse.json` | `github.com/ClickHouse/clickhouse-go/v2` | `clickhousev2.json` |
| `go-openai.json` | `github.com/sashabaranov/go-openai` | `openai.json` |
| `golog.json` | `log` | `log.json` |
| `goredis.json` | `github.com/go-redis/redis/v8`, `github.com/redis/go-redis/v9` | `redis.json` |
| `k8s-client.json` | `k8s.io/client-go/tools/cache` | `k8s_client_go.json` |
| `langchain.json` | `github.com/tmc/langchaingo/llms/openai`, `github.com/tmc/langchaingo/llms/ollama`, `github.com/tmc/langchaingo/embeddings`, ... | `langchaingo.json` |
| `mqttserver.json` | `github.com/mochi-mqtt/server/v2` | `mqtt.json` |
| `otel.json` | `go.opentelemetry.io/otel/baggage`, `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/trace`, ... | `base.json` |
| `otelmetric.json` | `go.opentelemetry.io/otel` | `base.json` |
| `rocketmq-client-go.json` | `github.com/apache/rocketmq-client-go/v2/producer`, `github.com/apache/rocketmq-client-go/v2/consumer` | `rocketmq.json` |
| `segmentio-kafka.json` | `github.com/segmentio/kafka-go` | `kafka.json` |

### 2) Fully Missing (No ImportPath Overlap)

| Missing Rule File | Example ImportPath(s) |
|---|---|
| `adk.json` | `google.golang.org/adk/internal/telemetry` |
| `ants.json` | `github.com/panjf2000/ants/v2`, `github.com/panjf2000/ants` |
| `confluent.json` | `github.com/confluentinc/confluent-kafka-go/kafka`, `github.com/confluentinc/confluent-kafka-go/v2/kafka` |
| `contrib.json` | `github.com/aliyun-arms/go-sdk/armscontext` |
| `cron.json` | `github.com/robfig/cron/v3` |
| `datatester.json` | `github.com/volcengine/datatester-go-sdk/event` |
| `deepseek.json` | `github.com/cohesion-org/deepseek-go` |
| `dify.json` | `github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel`, `github.com/langgenius/dify-plugin-daemon/internal/service`, `github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation`, ... |
| `fasthttproute.json` | `github.com/fasthttp/router` |
| `fc.json` | `github.com/aliyun/fc-runtime-go-sdk/fc` |
| `genai.json` | `google.golang.org/genai` |
| `gozerolog.json` | `github.com/zeromicro/go-zero/core/logx` |
| `ibm-kafka.json` | `github.com/IBM/sarama` |
| `imroc-req.json` | `github.com/imroc/req/v3` |
| `k8s-controller-client.json` | `sigs.k8s.io/controller-runtime/pkg/client` |
| `k8s-controller.json` | `sigs.k8s.io/controller-runtime/pkg/builder`, `sigs.k8s.io/controller-runtime/pkg/cache` |
| `new-api.json` | `github.com/QuantumNous/new-api/relay`, `github.com/QuantumNous/new-api/relay/common`, `github.com/QuantumNous/new-api/service`, ... |
| `opentracing.json` | `github.com/opentracing/opentracing-go`, `github.com/uber/jaeger-client-go` |
| `qoder_openai_proxy.json` | `gitlab.alibaba-inc.com/cosy/openai-proxy/pkg/chatter` |
| `rocketmq-clients.json` | `github.com/apache/rocketmq-clients/golang/v5` |
| `routine.json` | `github.com/timandy/routine` |
| `shopify-kafka.json` | `github.com/Shopify/sarama` |
| `streadway.json` | `github.com/streadway/amqp` |
| `thrift.json` | `github.com/apache/thrift/lib/go/thrift` |
| `volcengine-arkruntime.json` | `github.com/volcengine/volcengine-go-sdk/service/arkruntime`, `github.com/volcengine/volcengine-go-sdk/service/arkruntime/utils` |
| `xxljob.json` | `github.com/xxl-job/xxl-job-executor-go` |

## Raw Missing File List

- `adk.json`
- `ants.json`
- `clickhouse.json`
- `confluent.json`
- `contrib.json`
- `cron.json`
- `datatester.json`
- `deepseek.json`
- `dify.json`
- `fasthttproute.json`
- `fc.json`
- `genai.json`
- `go-openai.json`
- `golog.json`
- `goredis.json`
- `gozerolog.json`
- `ibm-kafka.json`
- `imroc-req.json`
- `k8s-client.json`
- `k8s-controller-client.json`
- `k8s-controller.json`
- `langchain.json`
- `mqttserver.json`
- `new-api.json`
- `opentracing.json`
- `otel.json`
- `otelmetric.json`
- `qoder_openai_proxy.json`
- `rocketmq-client-go.json`
- `rocketmq-clients.json`
- `routine.json`
- `segmentio-kafka.json`
- `shopify-kafka.json`
- `streadway.json`
- `thrift.json`
- `volcengine-arkruntime.json`
- `xxljob.json`
