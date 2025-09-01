## Fields of a rule definition

## Instrument a function
- `ImportPath`: The import path of the package that contains the function to be instrumented. e.g. `net/http`.
- `Dependencies`: is a list of additional dependencies that must be present for this rule to be applied. All dependencies must exist in the project. e.g. `"k8s.io/apimachinery"
- `Function`: The name of the function to be instrumented, it could be a regular expression to match multiple functions. e.g. `.*` matches all functions in the package, `.*ServeHTTP` matches all functions whose name ends with `ServeHTTP`, and so on.
- `ReceiverType`: The type of the receiver of the function to be instrumented, it could be a regular expression as well. e.g. `.*` matches all receiver types in the package, even if the function has no receiver, `.*` still matches it. `.*http.Request` matches all functions whose receiver type is `http.Request`, `\\*Client` matches all functions whose receiver type is `*Client`, and so on.
- `OnEnter`: The name of the function to be called when the instrumented function is called. e.g. `clientOnEnter`.
- `OnExit`: The name of the function to be called when the instrumented function returns. e.g. `clientOnExit`.
- `Order`: The order of the probe code in the instrumented function. e.g. `0`, `1`, `2`.
- `Path`: The path to the directory containing the probe code. The path can be either go module url or local file system path, e.g. `github.com/foo/bar` or `/path/to/probe/code`.
- `Version`: The version of the package that contains the function to be instrumented. e.g. `[1.0.0,1.1.0)`, the version range is `[1.0.0,1.1.0)`, which means the version is greater than or equal to `1.0.0` and less than `1.1.0`.

> ![TIP]
> You can use ".*" of both `Function` and `ReceiverType` to match all functions and all receiver types in the specific package.

## Add a new file during compiling package
- `ImportPath`: The import path of the package that contains the function to be instrumented.
- `FileName` : The name of the file to be added.
- `Path`: The path to the directory containing the probe code.
- `Replace`: Replace the file if it already exists, default is `false`.

## Add a new field to a struct
- `ImportPath`: The import path of the package that contains the struct to be instrumented.
- `StructType`: The name of the struct to be instrumented.
- `FieldName`: The name of the field to be added.
- `FieldType`: The type of the field to be added.
