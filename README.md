![](docs/public/anim-logo.svg)

[![](https://shields.io/badge/-Docs-blue?logo=readthedocs)](https://alibaba.github.io/loongsuite-go-agent/)  &nbsp;
[![](https://shields.io/badge/-商业版-blue?logo=alibabacloud)](https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications) &nbsp;

**Loongsuite Go Agent** provides an automatic solution for Golang applications that want to
leverage OpenTelemetry to enable effective observability. No code changes are
required in the target application, the instrumentation is done at compile
time. Simply adding `otel` prefix to `go build` to get started :rocket:

# Installation

### Prebuilt Binaries

- [![Download](https://shields.io/badge/-Linux_AMD64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-amd64)
- [![Download](https://shields.io/badge/-Linux_ARM64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-arm64)
- [![Download](https://shields.io/badge/-MacOS_AMD64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-amd64)
- [![Download](https://shields.io/badge/-MacOS_ARM64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-arm64)
- [![Download](https://shields.io/badge/-Windows_AMD64-blue?logo=wine)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-windows-amd64.exe)

**This is the recommended way to install the tool.**

### Install via Bash
For Linux and MacOS users, the following script will install `otel` in `/usr/local/bin/otel` by default:
```bash
$ sudo curl -fsSL https://cdn.jsdelivr.net/gh/alibaba/loongsuite-go-agent@main/install.sh | sudo bash
```

### Build from Source

```bash
$ make         # build only
$ make install # build and install
```

# Getting Started

Make sure the tool is installed:
```bash
$ # You may use "otel-linux-amd64" instead of "otel"
$ otel version
```

Just adding `otel` prefix to `go build` to build your project:

```bash
$ otel go build
$ otel go build -o app cmd/app
$ otel go build -gcflags="-m" cmd/app
```

That's the whole process! The tool will automatically instrument your code with OpenTelemetry, and you can start to observe your application. :telescope:

The detailed usage of `otel` tool can be found in [**Usage**](./docs/user/config.md).

> [!NOTE]
> If you find any compilation failures while `go build` works, it's likely a bug.
> Please feel free to file a bug
> at [GitHub Issues](https://github.com/alibaba/loongsuite-go-agent/issues)
> to help us enhance this project.

# Examples

- [demo](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/demo) - End-to-end example with OpenTelemetry tracing and metrics
- [zap logging](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/log) - Auto-instrumentation for `github.com/uber-go/zap` logging
- [benchmark](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/benchmark) - Performance testing and overhead measurement
- [sql injection](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/sqlinject) - Custom code injection for SQL injection detection
- [nethttp](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/nethttp) - HTTP monitoring with request/response instrumentation
- [kratos-demo](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/kratos-demo) - Integration with the Kratos framework
- [kafka-demo](https://github.com/alibaba/loongsuite-go-agent/tree/main/example/kafka-demo) - Kafka Consumer Message monitoring

# Supported Libraries
<details>
 <summary>List of Supported Libraries</summary>

## 数据库
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| database/sql   | https://pkg.go.dev/database/sql                             | -               | -            |
| gorm           | https://github.com/go-gorm/gorm                             | v1.22.0         | v1.25.9       |
| sqlx           | https://github.com/jmoiron/sqlx                             | v1.3.0          | v1.4.0        |
| gopg           | https://github.com/go-pg/pg                                 | v10.10.0        | v10.14.0      |
| mongodb        | https://github.com/mongodb/mongo-go-driver                  | v1.11.1         | v1.15.1       |
| elasticsearch  | https://github.com/elastic/go-elasticsearch                 | v8.4.0          | v8.15.0       |

## 缓存
| Library          | Repository Url                                           | Min Version     | Max Version |
|------------------|----------------------------------------------------------|-----------------|-------------|
| redis (go-redis) | https://github.com/redis/go-redis                        | v9.0.5          | v9.5.1      |
| redis v8         | https://github.com/go-redis/redis/v8                     | v8.11.0         | v8.11.5     |
| redigo           | https://github.com/gomodule/redigo                       | v1.9.0          | v1.9.3      |
| rueidis          | https://github.com/redis/rueidis                         | v1.0.30         | -           |

## 消息队列
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| rocketmq        | https://github.com/apache/rocketmq-client-go/v2             | v2.0.0          | -            |
| amqp091         | https://github.com/rabbitmq/amqp091-go                      | v1.10.0         | -            |
| segmentio/kafka-go| https://github.com/segmentio/kafka-go                     | v0.4.0          | -            |

## RPC/通信框架
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| grpc            | https://google.golang.org/grpc                              | v1.44.0         | -            |
| dubbo-go        | https://github.com/apache/dubbo-go                          | v3.3.0          | -            |
| kitex           | https://github.com/cloudwego/kitex                          | v0.5.1          | -            |
| kratos          | https://github.com/go-kratos/kratos                         | v2.6.3          | -            |
| go-micro        | https://github.com/micro/go-micro                           | v5.0.0          | v5.3.0        |
| trpc-go         | https://github.com/trpc-group/trpc-go                       | v1.0.0          | -            |

## HTTP/Web 框架
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| net/http        | https://pkg.go.dev/net/http                                 | -               | -            |
| echo            | https://github.com/labstack/echo                            | v4.0.0          | -            |
| gin             | https://github.com/gin-gonic/gin                            | v1.7.0          | v1.10.1       |
| fiber           | https://github.com/gofiber/fiber                            | v2.43.0         | v2.52.9       |
| fasthttp        | https://github.com/valyala/fasthttp                         | v1.45.0         | v1.65.0       |
| gorilla/mux     | https://github.com/gorilla/mux                              | v1.3.0          | v1.8.1        |
| iris            | https://github.com/kataras/iris                             | v12.2.0         | v12.2.11      |
| hertz           | https://github.com/cloudwego/hertz                          | v0.8.0          | -            |
| go-restful      | https://github.com/emicklei/go-restful                      | v3.7.0          | v3.12.1       |
| gorestful/v3    | https://github.com/emicklei/go-restful/v3                   | v3.7.0          | v3.12.1       |

## 配置/注册中心
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| nacos           | https://github.com/nacos-group/nacos-sdk-go/v2              | v2.0.0          | v2.2.9        |
| k8s client-go   | https://github.com/kubernetes/client-go                     | v0.33.3         | -            |

## 日志
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| log             | https://pkg.go.dev/log                                      | -               | -            |
| zap             | https://github.com/uber-go/zap                              | v1.20.0         | v1.27.0       |
| logrus          | https://github.com/sirupsen/logrus                          | v1.5.0          | v1.9.3        |
| zerolog         | https://github.com/rs/zerolog                               | v1.10.0         | v1.33.0       |
| go-kit/log      | https://github.com/go-kit/log                               | v0.1.0          | v0.2.1        |

## AI/LLM/向量/大模型
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| langchaingo     | https://github.com/tmc/langchaingo                          | v0.1.13         | -            |
| ollama          | https://github.com/ollama/ollama                            | v0.3.14         | -            |
| eino            | https://github.com/cloudwego/eino                           | v0.3.51         | -            |

## 限流/熔断
| Library         | Repository Url                                               | Min Version     | Max Version   |
|----------------|-------------------------------------------------------------|-----------------|--------------|
| sentinel        | https://github.com/alibaba/sentinel-golang                  | v1.0.4          | -            |

</details>

We are progressively open-sourcing the libraries we have supported, and your contributions are <kbd>Very Welcome</kbd>

> [!IMPORTANT]
> The framework you expected is not in the list? Don't worry, you can easily inject your code into any frameworks/libraries that are not officially supported.
>
> Please refer to [this document](./docs/dev/overview.md) to get started.

# Community

We are looking forward to your feedback and suggestions. You can join
our [DingTalk group](https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11)
to engage with us.

| DingTalk | Star History |
| :---: | :---: |
| <img src="./docs/public/dingtalk.png" height="200" /> | <img src="https://api.star-history.com/svg?repos=alibaba/loongsuite-go-agent&type=Date" height="200" /> |