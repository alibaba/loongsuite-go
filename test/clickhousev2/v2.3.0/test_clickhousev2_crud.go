package main

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"log"
	"os"
)

var con driver.Conn

type User struct {
	ID   uint
	Name string
	Age  uint8
}

func TestExec() {
	if err := con.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestAsyncInsert() {
	if err := con.AsyncInsert(context.Background(), `INSERT INTO users VALUES ('1', '1', 1)`, true); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestSelect() {
	var user User
	if err := con.Select(context.Background(), &user, `SELECT * FROM users WHERE id = ?`, 1); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestQuery() {
	if _, err := con.Query(context.Background(), `SELECT * FROM users WHERE id = ?`, 1); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestQueryRow() {
	if row := con.QueryRow(context.Background(), `SELECT * FROM users WHERE id = ?`, 1); row != nil && row.Err() != nil {
		log.Printf("%v\n", row.Err())
	}
}

func TestPrepareBatch() {
	batch, err := con.PrepareBatch(context.Background(), `SELECT * FROM users WHERE id = 1`)
	if err != nil {
		log.Printf("%v\n", err)
	}
	if err = batch.Send(); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestSelectVersion() {
	if _, err := con.ServerVersion(); err != nil {
		log.Printf("%v\n", err)
	}
}

func TestPing() {
	if err := con.Ping(context.Background()); err != nil {
		log.Printf("%v\n", err)
	}
}

func main() {
	tmpCon, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:" + os.Getenv("CLICKHOUSE_PORT")},
	})
	if err != nil {
		log.Fatalf("open connection fail, err: %v", err)
	}
	con = tmpCon
	TestExec()
	TestAsyncInsert()
	TestSelect()
	TestQuery()
	TestQueryRow()
	TestPrepareBatch()
	TestSelectVersion()
	TestPing()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "EXEC", "clickhouse", "127.0.0.1", "CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)", "EXEC", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "EXEC", "clickhouse", "127.0.0.1", "ping", "EXEC", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "SELECT", "clickhouse", "127.0.0.1", "ping", "SELECT", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "QUERY", "clickhouse", "127.0.0.1", "", "QUERY", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "QUERY_ROW", "clickhouse", "127.0.0.1", "START TRANSACTION", "QUERY_ROW", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "PREPARE_BATCH", "clickhouse", "127.0.0.1", "", "PREPARE_BATCH", "", nil)
		verifier.VerifyDbAttributes(stubs[6][0], "SELECT_VERSION", "clickhouse", "127.0.0.1", "", "SELECT_VERSION", "", nil)
		verifier.VerifyDbAttributes(stubs[7][0], "PING", "clickhouse", "127.0.0.1", "", "PING", "", nil)
	}, 1)
}
