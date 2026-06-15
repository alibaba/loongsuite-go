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

package databasesql

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type databaseSqlAttrsGetter struct {
}

func (d databaseSqlAttrsGetter) GetSystem(request databaseSqlRequest) string {
	switch request.driverName {
	case "mysql":
		return "mysql"
	case "postgres":
		fallthrough
	case "postgresql":
		return "postgresql"
	}
	return "database"
}

func (d databaseSqlAttrsGetter) GetServerAddress(request databaseSqlRequest) string {
	return request.endpoint
}

func (d databaseSqlAttrsGetter) GetStatement(request databaseSqlRequest) string {
	// Fetch metadata along with the SQL
	// Retrieve db collection only, not the params which is an experimental feature that will introduce some overhead.
	extractSQLMetadata(request)
	return request.sql
}

func (d databaseSqlAttrsGetter) GetOperation(request databaseSqlRequest) string {
	return request.opType
}

func (d databaseSqlAttrsGetter) GetCollection(request databaseSqlRequest) string {
	return getCollection(request.sql)
}

func (d databaseSqlAttrsGetter) GetParameters(request databaseSqlRequest) []any {
	if len(request.params) > 0 {
		// Prepared statement with parameter binding
		return request.params
	}

	// If it's not a prepared stmt with parameter binding, parse the raw sql
	return getParams(request.sql)
}

func (d databaseSqlAttrsGetter) GetDbNamespace(request databaseSqlRequest) string {
	return ""
}

func (d databaseSqlAttrsGetter) GetBatchSize(request databaseSqlRequest) int {
	return 0
}

func BuildDatabaseSqlOtelInstrumenter() instrumenter.Instrumenter[databaseSqlRequest, any] {
	builder := instrumenter.Builder[databaseSqlRequest, any]{}
	getter := databaseSqlAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[databaseSqlRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[databaseSqlRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[databaseSqlRequest, any, db.DbClientAttrsGetter[databaseSqlRequest]]{Base: db.DbClientCommonAttrsExtractor[databaseSqlRequest, any, db.DbClientAttrsGetter[databaseSqlRequest]]{Getter: getter}}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.DATABASE_SQL_SCOPE_NAME,
			Version: version.Tag,
		}).AddOperationListeners(db.DbClientMetrics("database.sql")).
		BuildInstrumenter()
}
