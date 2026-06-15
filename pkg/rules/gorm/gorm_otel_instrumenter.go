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

package gorm

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type gormAttrsGetter struct {
}

func (g gormAttrsGetter) GetSystem(gormRequest gormRequest) string {
	return gormRequest.System
}

func (g gormAttrsGetter) GetServerAddress(gormRequest gormRequest) string {
	return gormRequest.Endpoint
}

func (g gormAttrsGetter) GetStatement(gormRequest gormRequest) string {
	return ""
}

func (e gormAttrsGetter) GetCollection(gormRequest gormRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (g gormAttrsGetter) GetOperation(gormRequest gormRequest) string {
	return gormRequest.Operation
}

func (g gormAttrsGetter) GetParameters(gormRequest gormRequest) []any {
	return nil
}

func (g gormAttrsGetter) GetDbNamespace(gormRequest gormRequest) string {
	return ""
}

func (g gormAttrsGetter) GetBatchSize(gormRequest gormRequest) int {
	return 0
}

func BuildGormInstrumenter() instrumenter.Instrumenter[gormRequest, interface{}] {
	builder := instrumenter.Builder[gormRequest, interface{}]{}
	getter := gormAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[gormRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[gormRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[gormRequest, any, gormAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[gormRequest, any, gormAttrsGetter]{Getter: getter}}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GORM_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
