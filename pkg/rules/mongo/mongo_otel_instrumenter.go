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

package mongo

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"
)

type mongoAttrsGetter struct {
}

func (m mongoAttrsGetter) GetSystem(request mongoRequest) string {
	return "mongodb"
}

func (m mongoAttrsGetter) GetServerAddress(request mongoRequest) string {
	return request.Host
}

func (m mongoAttrsGetter) GetStatement(request mongoRequest) string {
	return request.CommandName
}

func (m mongoAttrsGetter) GetCollection(request mongoRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (m mongoAttrsGetter) GetOperation(request mongoRequest) string {
	return request.CommandName
}

func (m mongoAttrsGetter) GetParameters(request mongoRequest) []any {
	return nil
}

func (m mongoAttrsGetter) GetBatchSize(request mongoRequest) int {
	return 0
}

func (m mongoAttrsGetter) GetDbNamespace(request mongoRequest) string {
	return ""
}

type mongoSpanNameExtractor struct {
}

type mongoSpanKindExtractor struct {
}

func (m *mongoSpanKindExtractor) Extract(request mongoRequest) trace.SpanKind {
	return trace.SpanKindClient
}

func (m *mongoSpanNameExtractor) Extract(request mongoRequest) string {
	return request.CommandName
}

func BuildMongoOtelInstrumenter() instrumenter.Instrumenter[mongoRequest, interface{}] {
	builder := instrumenter.Builder[mongoRequest, interface{}]{}
	return builder.Init().SetSpanNameExtractor(&mongoSpanNameExtractor{}).
		AddOperationListeners(db.DbClientMetrics("nosql.mongo")).
		SetSpanKindExtractor(&mongoSpanKindExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MONGO_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[mongoRequest, any, db.DbClientAttrsGetter[mongoRequest]]{Base: db.DbClientCommonAttrsExtractor[mongoRequest, any, db.DbClientAttrsGetter[mongoRequest]]{Getter: mongoAttrsGetter{}}}).
		BuildInstrumenter()
}
