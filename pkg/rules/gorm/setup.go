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
	"context"
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var contextKey = "otel-context"
var requestKey = "otel-request"

type gormInnerEnabler struct {
	enabled bool
}

func (g gormInnerEnabler) Enable() bool {
	return g.enabled
}

var gormEnabler = gormInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GORM_ENABLED") != "false"}

var gormInstrumenter = BuildGormInstrumenter()

//go:linkname afterGormOpen gorm.io/gorm.afterGormOpen
func afterGormOpen(call api.CallContext, db *gorm.DB, err error) {
	if !gormEnabler.Enable() {
		return
	}
	if err != nil || db == nil {
		return
	}
	// add the callback
	_ = db.Callback().Create().Before("gorm:create").Register("otel_create_create_span", beforeCallback("", "create"))
	_ = db.Callback().Query().Before("gorm:query").Register("otel_create_query_span", beforeCallback("", "query"))
	_ = db.Callback().Update().Before("gorm:update").Register("otel_create_update_span", beforeCallback("", "update"))
	_ = db.Callback().Delete().Before("gorm:delete").Register("otel_create_delete_span", beforeCallback("", "delete"))
	_ = db.Callback().Row().Before("gorm:row").Register("otel_create_row_span", beforeCallback("", "row"))
	_ = db.Callback().Raw().Before("gorm:raw").Register("otel_create_raw_span", beforeCallback("", "raw"))

	// after database operation
	_ = db.Callback().Create().After("gorm:create").Register("otel_end_create_span", afterCallback(""))
	_ = db.Callback().Query().After("gorm:query").Register("otel_end_query_span", afterCallback(""))
	_ = db.Callback().Update().After("gorm:update").Register("otel_end_update_span", afterCallback(""))
	_ = db.Callback().Delete().After("gorm:delete").Register("otel_end_delete_span", afterCallback(""))
	_ = db.Callback().Row().After("gorm:row").Register("otel_end_row_span", afterCallback(""))
	_ = db.Callback().Raw().After("gorm:raw").Register("otel_end_raw_span", afterCallback(""))
}

func beforeCallback(endpoint string, op string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		dbName, addr, system, user := getDbInfo(db.Config.Dialector)
		request := gormRequest{
			DbName:    dbName,
			Endpoint:  addr,
			Operation: op,
			User:      user,
			System:    system,
		}
		ctx := gormInstrumenter.Start(context.Background(), request)
		db.Set(contextKey, ctx)
		db.Set(requestKey, request)
	}
}

func afterCallback(endpoint string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		iCtx, ok := db.Get(contextKey)
		if !ok {
			return
		}
		ctx, ok := iCtx.(context.Context)
		if !ok {
			return
		}
		iRequest, ok := db.Get(requestKey)
		if !ok {
			return
		}
		request, ok := iRequest.(gormRequest)
		if !ok {
			return
		}
		gormInstrumenter.End(ctx, request, nil, db.Statement.Error)
	}
}

func getDbInfo(dial gorm.Dialector) (string, string, string, string) {
	// TODO: support other database
	res, ok := dial.(*mysql.Dialector)
	if !ok {
		return "", "", "", ""
	}
	if cfg, ok := res.Config.DbInfo.(*driver.Config); ok {
		return cfg.DBName, cfg.Addr, "mysql", cfg.User
	}
	cfg, err := driver.ParseDSN(res.Config.DSN)
	if err != nil {
		return "", "", "", ""
	}
	res.Config.DbInfo = cfg
	return cfg.DBName, cfg.Addr, "mysql", cfg.User
}
