module github.com/alibaba/loongsuite-go-agent/pkg/ruless/sqlx

go 1.23.0

replace github.com/alibaba/loongsuite-go-agent/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go-agent/pkg v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.3.0 // indirect
)
