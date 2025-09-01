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

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type sqlxAttrsGetter struct{}

func (g sqlxAttrsGetter) GetSystem(sqlxRequest sqlxRequest) string {
	return sqlxRequest.driverName
}

func (g sqlxAttrsGetter) GetServerAddress(sqlxRequest sqlxRequest) string {
	return sqlxRequest.endpoint
}

func (g sqlxAttrsGetter) GetStatement(sqlxRequest sqlxRequest) string {
	return sqlxRequest.statement
}

func (g sqlxAttrsGetter) GetCollection(_ sqlxRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (g sqlxAttrsGetter) GetOperation(sqlxRequest sqlxRequest) string {
	return sqlxRequest.opType
}

func (g sqlxAttrsGetter) GetParameters(sqlxRequest sqlxRequest) []any {
	return sqlxRequest.params
}

func (g sqlxAttrsGetter) GetDbNamespace(sqlxRequest sqlxRequest) string {
	return sqlxRequest.dbName
}

func (g sqlxAttrsGetter) GetBatchSize(_ sqlxRequest) int {
	return 0
}

func BuildSqlxInstrumenter() instrumenter.Instrumenter[sqlxRequest, interface{}] {
	builder := instrumenter.Builder[sqlxRequest, interface{}]{}
	getter := sqlxAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[sqlxRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[sqlxRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[sqlxRequest, any, sqlxAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[sqlxRequest, any, sqlxAttrsGetter]{Getter: getter}}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SQLX_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
