// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func dbTransaction() {
	ctx := context.Background()
	db, err := sql.Open("mysql",
		"test:test@tcp(127.0.0.1:"+os.Getenv("MYSQL_PORT")+")/test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS users`); err != nil {
		log.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`); err != nil {
		log.Fatal(err)
	}
	stmt, err := db.PrepareContext(ctx, `INSERT INTO users (id, name, age) VALUES ( ?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	_, err = stmt.ExecContext(ctx, "1", "bar", 11)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tx.Exec(`INSERT INTO users (id, name, age) VALUES ( ?, ?, ? )`, "2", "foobar", 24); err != nil {
		log.Fatal(err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, "foobar", "0"); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "DROP", "mysql", "127.0.0.1", "DROP TABLE IF EXISTS users", "DROP", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "CREATE", "mysql", "127.0.0.1", "CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)", "CREATE", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "INSERT users", "mysql", "127.0.0.1", "INSERT INTO users (id, name, age) VALUES ( ?, ?, ?)", "INSERT", "users", []any{"1", "bar", 11})
		verifier.VerifyDbAttributes(stubs[3][0], "START", "mysql", "127.0.0.1", "START TRANSACTION", "START", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "INSERT users", "mysql", "127.0.0.1", "INSERT INTO users (id, name, age) VALUES ( ?, ?, ? )", "INSERT", "users", []any{"2", "foobar", 24})
		verifier.VerifyDbAttributes(stubs[5][0], "UPDATE users", "mysql", "127.0.0.1", "UPDATE users SET name = ? WHERE id = ?", "UPDATE", "users", []any{"foobar", "0"})
		verifier.VerifyDbAttributes(stubs[6][0], "COMMIT", "mysql", "127.0.0.1", "COMMIT", "COMMIT", "", nil)
	}, 7)
}

func main() {
	dbAccess()
	dbFetching()
	dbModify()
	dbPrepared()
	dbQuery()
	dbTransaction()
}
