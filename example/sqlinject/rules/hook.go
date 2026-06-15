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

package rules

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

func checkSqlInjection(query string) error {
	patterns := []string{"go", "build", ";", "/*", " or ", " and ", "'"}
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(query), pattern) {
			return errors.New("potential SQL injection detected")
		}
	}
	return nil
}

//go:linkname sqlQueryOnEnter database/sql.sqlQueryOnEnter
func sqlQueryOnEnter(call api.CallContext, db *sql.DB, query string, args ...interface{}) {
	if err := checkSqlInjection(query); err != nil {
		log.Fatalf("sqlQueryOnEnter %v", err)
	}
}
