# 上下文传播

`loongsuite-go-agent`中的上下文传播受到[Apache-Skywalking](https://github.com/apache/skywalking-go)的启发。
OpenTelemetry中的上下文是一种用于在分布式系统中传播与跟踪相关信息的设计。基于上下文的传播，分布式服务（即Spans）可以链接在一起，形成一个完整的调用链（即Trace）。OpenTelemetry将与跟踪相关的信息保存在Golang的context.Context中，并要求用户正确传递context.Context。如果context.Context在调用链中没有正确传递，调用链将会中断。为了解决这个问题，当`loongsuite-go-agent`创建一个span时，`loongsuite-go-agent`会将其保存到Golang的协程结构（即GLS）中，当`loongsuite-go-agent`创建一个新的协程时，`loongsuite-go-agent`也会从当前协程中复制相应的数据结构。当`loongsuite-go-agent`稍后需要创建一个新的span时，`loongsuite-go-agent`会从GLS中查询最近创建的span作为父级，这样`loongsuite-go-agent`就有机会保护调用链的完整性。

Baggage是OpenTelemetry中的一个数据结构，用于在Trace中共享键值对。Baggage存储在context.Context中，并随context.Context一起传播。如果context.Context在调用链中没有正确传播，后续服务将无法读取Baggage。为了解决这个问题，当`loongsuite-go-agent`将baggage保存到context.Context时，`loongsuite-go-agent`也会将其保存到GLS中。当context.Context没有正确传递时，`loongsuite-go-agent`会尝试从GLS中读取baggage，这使得在这种情况下可以读取baggage。
