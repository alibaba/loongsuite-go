// Copyright (c) 2025 Alibaba Group Holding Ltd.
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
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"log"
	"os"
)

type User struct {
	Id   string
	Name string
	Age  int
}

var db *sqlx.DB

func TestCreateTable() {
	if _, err := db.NamedExec(`CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`, struct{}{}); err != nil {
		log.Printf("%v", err)
	}
}

func TestInsert() {
	if _, err := db.NamedExec("INSERT INTO users (id, name, age) VALUES ( :id, :name, :age)", &User{
		Id:   "a",
		Name: "a",
		Age:  0,
	}); err != nil {
		log.Printf("%v", err)
	}
}

func TestQuery() {
	user := &User{}
	if err := db.Get(user, "select id, name from users where id = $1", "a"); err != nil {
		log.Printf("%v", err)
	}
}

func TestUpdate() {
	if _, err := db.NamedExec("UPDATE users set name = :name where id = :id", &User{
		Id:   "a",
		Name: "b",
	}); err != nil {
		log.Printf("%v", err)
	}
}

func TestDelete() {
	if _, err := db.NamedExec("delete from users where id = :id", &User{
		Id: "a",
	}); err != nil {
		log.Printf("%v", err)
	}
}

func TestDropTable() {
	if _, err := db.NamedExec("DROP TABLE IF EXISTS users", struct{}{}); err != nil {
		log.Printf("%v", err)
	}
}

func main() {
	host := "localhost:" + os.Getenv("MYSQL_PORT")
	mysqlDb, err := sqlx.Connect("mysql", "test:test@tcp("+host+")/test")
	if err != nil {
		panic(err)
	}
	db = mysqlDb
	defer db.Close()
	TestCreateTable()
	TestInsert()
	TestQuery()
	TestUpdate()
	TestDelete()
	TestDropTable()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "ping", "mysql", host, "ping", "ping", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "CREATE TABLE", "mysql", host, "CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)", "CREATE TABLE", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "INSERT", "mysql", host, "INSERT INTO users (id, name, age) VALUES ( :id, :name, :age)", "INSERT", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "SELECT", "mysql", host, "select id, name from users where id = $1", "SELECT", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "UPDATE", "mysql", host, "UPDATE users set name = :name where id = :id", "UPDATE", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "DELETE", "mysql", host, "delete from users where id = :id", "DELETE", "", nil)
		verifier.VerifyDbAttributes(stubs[6][0], "DROP TABLE", "mysql", host, "DROP TABLE IF EXISTS users", "DROP TABLE", "", nil)
	}, 1)
}
