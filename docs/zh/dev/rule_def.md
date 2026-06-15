# Hook规则的类型

## 插桩一个函数
- `ImportPath`: 包含要插桩的函数的包的导入路径。例如 `net/http`。
- `Dependencies`: 应用此规则必须存在的附加依赖项列表。所有依赖项必须存在于项目中。例如 `"k8s.io/apimachinery"`。
- `Function`: 要插桩的函数的名称，可以是正则表达式以匹配多个函数。例如 `.*` 匹配包中的所有函数，`.*ServeHTTP` 匹配名称以 `ServeHTTP` 结尾的所有函数，依此类推。
- `ReceiverType`: 要插桩的函数的接收器类型，也可以是正则表达式。例如 `.*` 匹配包中的所有接收器类型，即使函数没有接收器，`.*` 仍然匹配它。`.*http.Request` 匹配接收器类型为 `http.Request` 的所有函数，`\\*Client` 匹配接收器类型为 `*Client` 的所有函数，依此类推。
- `OnEnter`: 当被插桩的函数被调用时要调用的函数的名称。例如 `clientOnEnter`。
- `OnExit`: 当被插桩的函数返回时要调用的函数的名称。例如 `clientOnExit`。
- `Path`: 包含探针代码的目录的路径。路径可以是go模块url或本地文件系统路径，例如 `github.com/foo/bar` 或 `/path/to/probe/code`。
- `Version`: 包含要插桩的函数的包的版本。例如 `[1.0.0,1.1.0)`，版本范围是 `[1.0.0,1.1.0)`，这意味着版本大于或等于 `1.0.0` 且小于 `1.1.0`。

> ![TIP]
> 您可以同时使用 `Function` 和 `ReceiverType` 的 ".*" 来匹配特定包中的所有函数和所有接收器类型。

## 在编译包期间添加一个新文件
- `ImportPath`: 包含要插桩的函数的包的导入路径。
- `FileName` : 要添加的文件的名称。
- `Path`: 包含探针代码的目录的路径。
- `Replace`: 如果文件已存在，则替换它，默认为 `false`。

## 向结构体添加一个新字段
- `ImportPath`: 包含要插桩的结构体的包的导入路径。
- `StructType`: 要插桩的结构体的名称。
- `FieldName`: 要添加的字段的名称。
- `FieldType`: 要添加的字段的类型。
