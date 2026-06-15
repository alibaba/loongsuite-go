# 注册Hook规则
我们需要在`tool/data/rules`目录下添加一个以规则命名的JSON文件，比如`nethttp.json`，来注册这个规则：
```json
[{
  "Version": "[1.3.0,1.7.4)",
  "ImportPath": "github.com/gorilla/mux",
  "Function": "setCurrentRoute",
  "OnEnter": "muxRoute130OnEnter",
  "Path": "github.com/alibaba/loongsuite-go/pkg/rules/mux"
},...]
```

以`github.com/gorilla/mux`为例，这个条目声明了我们想在目标函数`setCurrentRoute`的开头注入我们的埋点函数`muxRoute130OnEnter`。埋点代码位于`github.com/alibaba/loongsuite-go/pkg/rules/mux`目录下，支持的mux版本是`[1.3.0,1.7.4)`。

更详细的字段定义，请参考[rule_def.md](rule_def.md)。
