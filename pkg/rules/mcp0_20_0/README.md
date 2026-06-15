Monitor the three methods: **beforeAny, onSuccess, and onError**. All existing hook methods will execute these three methods. beforeAny serves as the start of OpenTelemetry (OTel) tracing, while onSuccess or onError marks the end of OTel tracing.

Currently, the stdio client and server cannot be correlated.

The monitored events are as follows:

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
