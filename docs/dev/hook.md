# Write the Hook Code
We need to create a new plugin directory under pkg/rules/ and then write the plugin code, like this:

```go
package mux

import _ "unsafe"
import "github.com/alibaba/loongsuite-go-agent/pkg/api"
import mux "github.com/gorilla/mux"

//go:linkname muxRoute130OnEnter github.com/gorilla/mux.muxRoute130OnEnter
func muxRoute130OnEnter(call api.CallContext, req *http.Request, route interface{}) {
    ...
}
```
There's no magic here — it's just regular Go code. A few interesting points:

- The hook function muxRoute130OnEnter must be annotated with the go:linkname directive.
- The first parameter of the hook function must be of type api.CallContext, and the remaining parameters should match those of the target function:
  - If the target function is `func foo(a int, b string, c float) (d string, e error)`, then the onEnter hook function should be `func hook(call api.CallContext, a int, b string, c float)`
  - If the target function is `func foo(a int, b string, c float) (d string, e error)`, then the onExit hook function should be `func hook(call api.CallContext, d string, e error)`
  - If you need to modify the parameters or return values of the target function, you can use `CallContext.SetParam()` or `CallContext.SetReturnVal()`

We need more documentation explaining all aspects of writing plugin code. For now, the best way is to refer to other plugin implementations, such as `pkg/rules/mux` or any other existing plugin.