# Use extensions to detect sql injection

The document demonstrates how to use extensions to detect sql injection.

## Step1: Compile the target binary with otel
Use `otel` to build the binary with `config.json`:
```
$ cd example/sqlinject
$ ../../otel set -rule=config.json
$ ../../otel go build .
```
Please make sure `otel` is correctly installed/built.

## Step2: Run the binary compiled by otel
```shell
$ docker run -d -p 3306:3306 -p 33060:33060 -e MYSQL_USER=test -e MYSQL_PASSWORD=test -e MYSQL_DATABASE=test -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:8.0.36
$ ./sqlinjectdemo
```
And the result will be:
```shell
2024/11/04 21:24:55 sqlQueryOnEnter potential SQL injection detected
```
Custom hook function is correctly injected.

