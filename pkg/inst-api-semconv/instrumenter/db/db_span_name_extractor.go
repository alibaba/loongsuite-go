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

package db

type DBSpanNameExtractor[REQUEST any] struct {
	Getter DbClientAttrsGetter[REQUEST]
}

// ref: https://opentelemetry.io/docs/specs/semconv/database/database-spans/#name
func (d *DBSpanNameExtractor[REQUEST]) Extract(request REQUEST) string {
	operation := d.Getter.GetOperation(request)
	target := d.Getter.GetCollection(request)
	system := d.Getter.GetSystem(request)

	if operation != "" && target != "" {
		return operation + " " + target
	}

	if operation != "" {
		return operation
	}
	// If neither {db.operation.name} nor {target} are available, span name SHOULD be {db.system}.
	if system != "" {
		return system
	}
	
	return "DB"
}