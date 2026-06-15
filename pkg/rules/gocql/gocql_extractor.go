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

package gocql

import "strings"

const (
	defaultOp        = "QUERY"
	createKeyspaceOp = "CREATE KEYSPACE"
	createTableOp    = "CREATE TABLE"
	insertOp         = "INSERT"
	selectOp         = "SELECT"
	updateOp         = "UPDATE"
	deleteOp         = "DELETE"
	dropTableOp      = "DROP TABLE"
	dropKeySpaceOp   = "DROP KEYSPACE"
)

var opColl = []string{createKeyspaceOp, createTableOp, insertOp, selectOp, updateOp, deleteOp, dropTableOp, dropKeySpaceOp}

func extractOpType(statement string) string {
	s := strings.TrimSpace(strings.ToUpper(statement))
	for _, op := range opColl {
		if strings.HasPrefix(s, op) {
			return op
		}
	}
	return defaultOp
}
