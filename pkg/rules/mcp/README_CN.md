监听**beforeAny，onSuccess，onError**三个方法。现有hook方法都会执行这三个个方法。beforeAny作为otel起始，onSuccess或onError作为otel结束。

目前stdio client与sever无法关联

监听事件如下：

```
MethodInitialize MCPMethod = "initialize"
MethodResourcesList MCPMethod = "resources/list"
MethodResourcesTemplatesList MCPMethod = "resources/templates/list"
MethodResourcesRead MCPMethod = "resources/read"
MethodPromptsList MCPMethod = "prompts/list"
MethodPromptsGet MCPMethod = "prompts/get"
MethodToolsList MCPMethod = "tools/list"
MethodToolsCall MCPMethod = "tools/call"
```
