# Adding a New Rule
This document will briefly describe how to add a new plugin to the official repository — that is, how to inject instrumentation code into a new third-party library.

## 1. Registering the New Rule
We need to add a JSON file named after the rule, such as nethttp.json, in the tool/data/rules directory to register this rule:
```json
[{
  "Version": "[1.3.0,1.7.4)",
  "ImportPath": "github.com/gorilla/mux",
  "Function": "setCurrentRoute",
  "OnEnter": "muxRoute130OnEnter",
  "Path": "github.com/alibaba/loongsuite-go-agent/pkg/rules/mux"
},...]
```

Taking `github.com/gorilla/mux` as an example, this entry declares that we want to inject our instrumentation function `muxRoute130OnEnter` at the beginning of the target function `setCurrentRoute`. The instrumentation code is located under the directory `github.com/alibaba/loongsuite-go-agent/pkg/rules/mux`, and the supported versions of mux are `[1.3.0,1.7.4)`.

For more detailed field definitions, please refer to [rule_def.md](rule_def.md).

## 2. Writing the Plugin Code
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

## 3. Testing the Plugin
Please refer to [how-to-write-tests-for-plugins.md](how-to-write-tests-for-plugins.md) for details.
