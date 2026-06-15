# Compilation Time
When using our `otel` tool, there will be a noticeable increase in compilation time. The main reason is that we introduce new dependencies and execute `go mod tidy` to fetch these dependencies, which consumes time depending on the network bandwidth. On the other hand, we inject code into the standard library, and the injected code may confuse users. For this reason, we enforce a full recompilation every time instead of incremental compilation. The combined result of these factors is an increase in compilation time. We plan to focus on optimizing this aspect in the future, but as of now, there has been no progress in this area.

When using our automatic instrumentation tool,
two additional phases are added before the above steps: **Preprocessing** and **Instrument**.
The total compilation time is the sum of the time taken by these two phases and
the time taken by the original compilation process. In general, `~92.8%` of the 
total compilation time is increased due to these two phases.