# 与手动埋点集成

自动埋点在大多数场景下已经能满足我们的需求，但手动埋点允许开发者对他们的项目有更大的控制权。

### 自动埋点

基于`example/demo`，自动埋点会生成一个trace，其中HTTP服务作为根span，Redis和MySQL操作作为子span。

![](../../public/auto_instr_jaeger.png)

### 结合手动埋点

手动埋点使我们能够捕获特定的遥测数据。在`example/demo/pkg/http.go`中，我们可以向包装数据库操作的`traceService()`函数添加一个手动span。

```go
var tracer = otel.Tracer("otel-manual-instr")

func traceService(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "db init")
	defer span.End()
    
    ...
}
```

在Jaeger中生成的trace如下。

![](../../public/manual_instr_jaeger.png)
