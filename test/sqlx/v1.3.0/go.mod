module sqlx/v1.3.0

go 1.23.0

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../../test/verifier

require (
	github.com/alibaba/loongsuite-go-agent/test/verifier v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.3.0
	go.opentelemetry.io/otel/sdk v1.35.0
	github.com/go-sql-driver/mysql v1.8.1
)