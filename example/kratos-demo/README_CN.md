# kratos-demo

## 项目简介 [English](./README.md)

这是一个在 Kratos 框架环境下使用 loongsuite-go-agent 的演示项目。

## 架构说明

本演示项目包含两个服务：

- **WeatherService**: HTTP 服务，监听端口 8080
- **MessageService**: gRPC 服务，监听端口 8081

### 服务调用流程

1. WeatherService 接收外部 HTTP 请求
2. WeatherService 通过 gRPC 调用 MessageService
3. MessageService 返回处理结果
4. WeatherService 将结果返回给客户端

### 可观测性

项目使用 loongsuite-go-agent 将链路追踪（Trace）和监控指标（Metric）导出到：
- **OpenTelemetry Collector**: 数据收集和处理
- **Jaeger**: 分布式链路追踪
- **Prometheus**: 监控指标存储

## 快速启动

### 运行服务

```shell
sh start.sh
```

### 测试 API

启动完成后，可以通过以下方式测试服务：

```shell
curl http://localhost:8080/v1/weather/{city}/message
```

**示例**：
```shell
curl http://localhost:8080/v1/weather/HangZhou/message
```

### 查看监控数据

- **Jaeger 链路追踪**: http://localhost:16686，发起请求后，刷新即可看到链路追踪数据。
- **Prometheus 监控指标**: http://localhost:9090，查询‘{exported_job="kratos-demo"}’即可看到kratos-demo的监控指标。

## 端口说明

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| WeatherService | 8080 | HTTP | 对外提供 REST API |
| MessageService | 8081 | gRPC | 内部服务调用 |
| Jaeger | 16686 | HTTP | 链路追踪 UI |
| Prometheus | 9090 | HTTP | 监控指标 UI |
