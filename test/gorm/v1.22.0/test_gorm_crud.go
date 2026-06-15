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
	"log"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID   uint
	Name string
	Age  uint8
}

func TestRaw() {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`).Error; err != nil {
		log.Printf("%v", err)
	}
}

func TestCreate() {
	user := User{Name: "opentelemetry", Age: 18}
	if err := db.Create(&user).Error; err != nil {
		log.Printf("%v", err)
	}
}

func TestQuery() {
	var user User
	if err := db.First(&user).Error; err != nil {
		log.Printf("%v", err)
	}
}

func TestRow() {
	var name string
	var age uint8
	row := db.Table("users").Where("name = ?", "opentelemetry").Select("name", "age").Row()
	row.Scan(&name, &age)
}

func TestUpdate() {
	tx := db.Model(&User{}).Where("name = ?", "opentelemetry").Update("name", "hello")
	if err := tx.Error; err != nil {
		log.Printf("%v", err)
	}
}

func TestDelete() {
	if err := db.Delete(&User{}, 1).Error; err != nil {
		log.Printf("%v", err)
	}
}

func main() {
	tmpDB, err := gorm.Open(mysql.Open("test:test@tcp(127.0.0.1:"+os.Getenv("MYSQL_PORT")+")/test"), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}
	db = tmpDB
	TestRaw()
	TestCreate()
	TestQuery()
	TestRow()
	TestUpdate()
	TestDelete()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "SELECT dual", "mysql", "127.0.0.1", "SELECT VERSION()", "SELECT", "dual", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "ping", "mysql", "127.0.0.1", "ping", "ping", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "raw", "mysql", "127.0.0.1", "", "raw", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "START", "mysql", "127.0.0.1", "START TRANSACTION", "START", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "create", "mysql", "127.0.0.1", "", "create", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "query", "mysql", "127.0.0.1", "", "query", "", nil)
		verifier.VerifyDbAttributes(stubs[6][0], "row", "mysql", "127.0.0.1", "", "row", "", nil)
		verifier.VerifyDbAttributes(stubs[7][0], "START", "mysql", "127.0.0.1", "START TRANSACTION", "START", "", nil)
		verifier.VerifyDbAttributes(stubs[8][0], "update", "mysql", "127.0.0.1", "", "update", "", nil)
		verifier.VerifyDbAttributes(stubs[9][0], "START", "mysql", "127.0.0.1", "START TRANSACTION", "START", "", nil)
		verifier.VerifyDbAttributes(stubs[10][0], "delete", "mysql", "127.0.0.1", "", "delete", "", nil)
	}, 1)
}
