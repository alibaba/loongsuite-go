module test/ollama

go 1.23.0

toolchain go1.24.1

require github.com/ollama/ollama v0.3.14

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
)

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg
