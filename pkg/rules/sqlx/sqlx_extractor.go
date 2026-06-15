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

package sqlx

import "strings"

var opTypes = []string{"CREATE TABLE", "DROP TABLE", "ALTER TABLE", "SELECT", "INSERT", "UPDATE", "DELETE"}

func extractOpType(statement string) string {
	upperStatement := strings.ToUpper(statement)
	for _, opType := range opTypes {
		if strings.Contains(upperStatement, opType) {
			return opType
		}
	}
	return ""
}
