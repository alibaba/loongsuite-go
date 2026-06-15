## benchmark

In this section, `loongsuite-go-agent` provides a simple benchmark demo to show its resource
consumption.

## benchmark environment

The benchmark is executed as the following picture:

```
 ________              ________              ________
|        |            |        |            |        |
|  PTS   |----------->|Consumer|----------->|Provider|
|________|            |________|            |________|
```

Benchmark uses [PTS](https://pts.console.aliyun.com/) to send requests to the consumer, consumer redirects the traffic
to the provider. The consumer and provider are both a simple `nethttp` Golang service.

All the services run on Alibaba Cloud [ACK](https://cs.console.aliyun.com/), The specifications of the machine are shown
below:

| MACHINE        | CPU | MEMORY |
|----------------|-----|--------|
| ecs.c7.6xlarge | 1C  | 4096MB |

After the consumer and provider are deployed, benchmark starts to send requests by using PTS, the PTS script yaml is
shown as below. Benchmark uses `1000` QPS as a benchmark to stress test both the consumer services with and without the
agent.

```yaml
---
Relations:
  - Disabled: false
    Id: "72WAL"
    Name: "loongsuite-go-agent"
    Nodes:
      - Config:
          accessId: "4ZU5L"
          beginStep: 1000
          checkPoints: [ ]
          endStep: 1000
          headers: [ ]
          method: "GET"
          nodeType: "CHAIN"
          postActions: [ ]
          protocol: "http"
          redirectCountLimit: 10
          url: "http://${consumerIp}:8080/echo"
        Name: "echo"
        NodeId: "1WS0L"
        NodeType: "chain"
        pressureNode: true
    RelationExecuteConfig:
      ExecuteCount: 0
      RelationExecuteType: "normal"
      relationExecuteType: "normal"
    relationExecuteConfig:
      $ref: "$.Relations[0].RelationExecuteConfig"
    relationTestConfig: { }
```

## benchmark report

|        | NO-AGENT | WITH-AGENT |
|--------|----------|------------|
| CPU    | 234M     | 325M       |
| Memory | 40MB     | 44MB       |
| RT     | 33ms     | 34ms       |

The CPU usage increases by about 9%(32.5%~23.4%), memory usage
experienced almost no growth, and the RT increases by approximately 1ms.