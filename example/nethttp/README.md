# Custom code injection: print http request and response header

The document demonstrates how to inject any code you want to the target binary.

## Step1: Compile the target binary with otel
Use `otel` to build the binary with `config.json`:
```
$ cd example/nethttp
$ ../../otel set -rule=config.json
$ ../../otel go build .
```
Please make sure `otel` is correctly installed/built.

## Step2: Run the binary compiled by otel
```shell
$ ./demo
```
And the result will be:
```shell
request header is  {"otel":["true"]}
response header is  {"Content-Type":["application/x-gzip"],"Date":["Wed, 06 Nov 2024 11:35:37 GMT"],"Server":["bfe"]}
```
Custom hook function is correctly injected.