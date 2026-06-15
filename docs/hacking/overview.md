Under normal circumstances, the `go build` command goes through the following main steps to compile a Golang application:

1. Source Code Parsing: The Golang compiler first parses the source code files and transforms them into an Abstract Syntax Tree (AST).
2. Type Checking: After parsing, type checking ensures that the code adheres to Golang's type system.
3. Semantic Analysis: This involves analyzing the semantics of the program, including variable definitions and usages, as well as package imports.
4. Compilation Optimization: The syntax tree is converted into an intermediate representation and various optimizations are performed to improve code execution efficiency.
5. Code Generation: Machine code for the target platform is generated.
6. Linking: Different packages and libraries are linked together to form a single executable file.

When using our automatic instrumentation tool, two additional phases are added before the above steps: **Preprocessing** and **Instrument**.

![](../public/workflow.png)

- `Preprocess`: Analyze dependencies and select rules that should be used later.
- `Instrument`: Generate code based on rules and inject new code into source code.
