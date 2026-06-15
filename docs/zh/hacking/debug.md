# 调试

## 1. 使用调试选项执行埋点

```bash
$ otel set -debug 
$ otel set -verbose
```

使用`-debug`编译选项时，该工具将编译一个未优化的二进制文件，同时保留所有生成的临时文件，例如调试日志和匹配的规则。您可以查看它们以了解该工具注入了什么样的代码。`-verbose`日志将向您显示该工具的详细过程。

## 2. 检查`.otel-build`目录

即使不使用`-debug`选项，该工具也会在`.otel-build`中保留必要的修改文件副本，其结构如下：

```shell
.otel-build
├── debug.log
├── instrument # 埋点代码，这正是我们注入的代码
│   ├── baggage
│   │   ├── otel_inst_file_context.go
│...
└── preprocess # 预处理阶段生成的临时文件
    ├── backups # 原始文件的备份，以备回滚
    ├── changed # 预处理阶段更改的文件
    ├── dry_run.log # 空运行日志
    └── matched_rules.json # 匹配的规则
```
