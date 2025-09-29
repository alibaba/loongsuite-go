# Registering the Hook Rule
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