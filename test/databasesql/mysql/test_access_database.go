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
	"database/sql"
	"log"
	"os"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func dbAccess() {
	db, err := sql.Open("mysql",
		"test:test@tcp(127.0.0.1:"+os.Getenv("MYSQL_PORT")+")/test")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "ping", "mysql", "127.0.0.1", "ping", "ping", "", nil)
	}, 1)
}
