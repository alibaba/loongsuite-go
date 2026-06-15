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

package gopg

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"os"
	_ "unsafe"
)

var requestKey = "otel-request"

type gopgInnerEnabler struct {
	enabled bool
}

func (g gopgInnerEnabler) Enable() bool {
	return g.enabled
}

var gopgEnabler = gopgInnerEnabler{
	enabled: os.Getenv("OTEL_INSTRUMENTATION_GOPG_ENABLED") != "false",
}

var gopgInstrumenter = BuildGopgInstrumenter()

type otelQueryHooker struct {
	db *pg.DB
}

type queryOperation interface {
	Operation() orm.QueryOp
}

func (o *otelQueryHooker) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	var query string
	var operation orm.QueryOp

	if v, ok := event.Query.(queryOperation); ok {
		operation = v.Operation()
	}
	if b, err := event.FormattedQuery(); err == nil {
		query = string(b)
	}
	request := gopgRequest{
		QueryOp:   operation,
		Statement: query,
		Addr:      o.db.Options().Addr,
		User:      o.db.Options().User,
		DbName:    o.db.Options().Database,
	}
	return context.WithValue(gopgInstrumenter.Start(ctx, request), requestKey, request), nil
}

func (o *otelQueryHooker) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	request := ctx.Value(requestKey).(gopgRequest)
	gopgInstrumenter.End(ctx, request, nil, event.Err)
	return nil
}

//go:linkname afterGopgConnect github.com/go-pg/pg/v10.afterGopgConnect
func afterGopgConnect(_ api.CallContext, db *pg.DB) {
	if !gopgEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	db.AddQueryHook(&otelQueryHooker{db: db})
}
