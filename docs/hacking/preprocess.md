# Preprocess Phase

In this phase, the tool analyzes third-party library dependencies in the user's project 
code and matches them against existing instrumentation rules to find appropriate rules. 
It also pre-configures the extra dependencies required by these rules.

Instrumentation rules precisely define which code needs to be injected into which 
version of which framework or standard library. Different types of instrumentation 
rules serve different purposes. The currently available types of instrumentation
rules include:

- InstFuncRule: Inject code at the entry and exit points of a method.
- InstStructRule: Modify a struct by adding a new field.
- InstFileRule: Add a new file to participate in the original compilation process.

Once all the preprocessing is complete, `go build -toolexec otel cmd/app` 
is called for compilation. The `-toolexec` parameter is the core of our automatic 
instrumentation, used to intercept the conventional build process and replace it
with a user-defined tool, allowing developers to customize the build process more 
flexibly. Here, `otel` is the automatic instrumentation tool,
which brings us to the Instrument phase.