# 编写Hook代码
我们需要在`pkg/rules/`下创建一个新的插件目录，然后编写插件代码，如下所示：

```go
package mux

import _ "unsafe"
import "github.com/alibaba/loongsuite-go/pkg/api"
import mux "github.com/gorilla/mux"

//go:linkname muxRoute130OnEnter github.com/gorilla/mux.muxRoute130OnEnter
func muxRoute130OnEnter(call api.CallContext, req *http.Request, route interface{}) {
    ...
}
```
这里没有什么魔法——它只是常规的Go代码。有几个有趣的点：

- hook函数`muxRoute130OnEnter`必须用`go:linkname`指令进行注解。
- hook函数的第一个参数必须是`api.CallContext`类型，其余参数应与目标函数的参数匹配：
  - 如果目标函数是`func foo(a int, b string, c float) (d string, e error)`，那么onEnter hook函数应该是`func hook(call api.CallContext, a int, b string, c float)`
  - 如果目标函数是`func foo(a int, b string, c float) (d string, e error)`，那么onExit hook函数应该是`func hook(call api.CallContext, d string, e error)`
  - 如果你需要修改目标函数的参数或返回值，你可以使用`CallContext.SetParam()`或`CallContext.SetReturnVal()`

我们需要更多的文档来解释编写插件代码的所有方面。目前，最好的方法是参考其他插件的实现，比如`pkg/rules/mux`或任何其他现有的插件。
