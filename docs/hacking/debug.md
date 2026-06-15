# Debugging

## 1. Perform instrumentation with debug options

```bash
$ otel set -debug 
$ otel set -verbose
```

When using the `-debug` compilation option, the tool will compile an unoptimized binary 
while retaining all generated temporary files, such as debug logs and matched rules. You can review 
them to understand what kind of code the tool is injecting. The `-verbose` log will show you the detailed process of the tool.

## 2. Check `.otel-build` directory

Even without using the `-debug` option, the tool will retain the necessary modified file copies in `.otel-build`, and its structure is as follows:

```shell
.otel-build
├── debug.log
├── instrument # instrumented code, which is exactly the code we injected
│   ├── baggage
│   │   ├── otel_inst_file_context.go
│...
└── preprocess # temporary files generated during preprocess phase
    ├── backups # backup of original files in case of rollback
    ├── changed # changed files during preprocess phase
    ├── dry_run.log # dry run log
    └── matched_rules.json # matched rules
```