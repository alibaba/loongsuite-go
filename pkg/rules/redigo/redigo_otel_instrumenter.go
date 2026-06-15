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

package redigo

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"strings"
	"time"
)

type redigoRequest struct {
	args      []interface{}
	endpoint  string
	cmd       string
	ctx       context.Context
	startTime time.Time
}

type redigoAttrsGetter struct {
}

func (m redigoAttrsGetter) GetSystem(request *redigoRequest) string {
	return "redis"
}

func (m redigoAttrsGetter) GetServerAddress(request *redigoRequest) string {
	return request.endpoint
}

func (m redigoAttrsGetter) GetStatement(request *redigoRequest) string {
	builder := strings.Builder{}
	builder.WriteString(request.cmd + " ")
	for _, arg := range request.args {
		builder.WriteString(fmt.Sprintf("%v ", arg))
	}
	return builder.String()
}

func (m redigoAttrsGetter) GetOperation(request *redigoRequest) string {
	return request.cmd
}

func (m redigoAttrsGetter) GetCollection(request *redigoRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (m redigoAttrsGetter) GetParameters(request *redigoRequest) []any {
	return nil
}

func (m redigoAttrsGetter) GetDbNamespace(request *redigoRequest) string {
	return ""
}

func (m redigoAttrsGetter) GetBatchSize(request *redigoRequest) int {
	return 0
}

func BuildRedigoInstrumenter() instrumenter.Instrumenter[*redigoRequest, interface{}] {
	builder := instrumenter.Builder[*redigoRequest, any]{}
	getter := redigoAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[*redigoRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*redigoRequest]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.REDIGO_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(db.DbClientMetrics("nosql.redigo")).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[*redigoRequest, any, db.DbClientAttrsGetter[*redigoRequest]]{Base: db.DbClientCommonAttrsExtractor[*redigoRequest, any, db.DbClientAttrsGetter[*redigoRequest]]{Getter: getter}}).
		BuildInstrumenter()
}
