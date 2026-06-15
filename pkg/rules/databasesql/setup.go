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
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
)

var databaseSqlInstrumenter = BuildDatabaseSqlOtelInstrumenter()

type dbSqlInnerEnabler struct {
	enabled bool
}

func (d dbSqlInnerEnabler) Enable() bool {
	return d.enabled
}

var dbSqlEnabler = dbSqlInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_DATABASESQL_ENABLED") != "false"}

const (
	cacheUpperBound = 1024
)

var sqlCache *SQLMetaCache

func init() {
	var err error
	sqlCache, err = NewSQLMetaCache(cacheUpperBound)
	if err != nil {
		log.Printf("failed to initialize SQL metadata cache: %v", err)
	}
}

//go:linkname beforeOpenInstrumentation database/sql.beforeOpenInstrumentation
func beforeOpenInstrumentation(call api.CallContext, driverName, dataSourceName string) {
	if !dbSqlEnabler.Enable() {
		return
	}
	addr, err := parseDSN(driverName, dataSourceName)
	if err != nil {
		log.Printf("failed to parse dsn: %v", err)
	}
	call.SetData(map[string]string{
		"endpoint": addr,
		"driver":   driverName,
		"dsn":      dataSourceName,
	})
}

//go:linkname afterOpenInstrumentation database/sql.afterOpenInstrumentation
func afterOpenInstrumentation(call api.CallContext, db *sql.DB, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	data, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	endpoint, ok := data["endpoint"]
	if ok {
		db.Endpoint = endpoint
	}
	driver, ok := data["driver"]
	if ok {
		db.DriverName = driver
	}
	dsn, ok := data["dsn"]
	if ok {
		db.DSN = dsn
	}
}

//go:linkname beforePingContextInstrumentation database/sql.beforePingContextInstrumentation
func beforePingContextInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	instrumentStart(call, ctx, "ping", "ping", db.Endpoint, db.DriverName, db.DSN)
}

//go:linkname afterPingContextInstrumentation database/sql.afterPingContextInstrumentation
func afterPingContextInstrumentation(call api.CallContext, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforePrepareContextInstrumentation database/sql.beforePrepareContextInstrumentation
func beforePrepareContextInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context, query string) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	call.SetData(map[string]string{
		"endpoint": db.Endpoint,
		"sql":      query,
		"driver":   db.DriverName,
		"dsn":      db.DSN,
	})
}

//go:linkname afterPrepareContextInstrumentation database/sql.afterPrepareContextInstrumentation
func afterPrepareContextInstrumentation(call api.CallContext, stmt *sql.Stmt, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	callDataMap, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	stmt.Data = map[string]string{
		"endpoint": callDataMap["endpoint"],
		"sql":      callDataMap["sql"],
		"driver":   callDataMap["driver"],
	}
	stmt.DSN = callDataMap["dsn"]
}

//go:linkname beforeExecContextInstrumentation database/sql.beforeExecContextInstrumentation
func beforeExecContextInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	instrumentStart(call, ctx, "exec", query, db.Endpoint, db.DriverName, db.DSN, args...)
}

//go:linkname afterExecContextInstrumentation database/sql.afterExecContextInstrumentation
func afterExecContextInstrumentation(call api.CallContext, result sql.Result, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeQueryContextInstrumentation database/sql.beforeQueryContextInstrumentation
func beforeQueryContextInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	instrumentStart(call, ctx, "query", query, db.Endpoint, db.DriverName, db.DSN, args...)
}

//go:linkname afterQueryContextInstrumentation database/sql.afterQueryContextInstrumentation
func afterQueryContextInstrumentation(call api.CallContext, rows *sql.Rows, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeTxInstrumentation database/sql.beforeTxInstrumentation
func beforeTxInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context, opts *sql.TxOptions) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	instrumentStart(call, ctx, "begin", "START TRANSACTION", db.Endpoint, db.DriverName, db.DSN)
}

//go:linkname afterTxInstrumentation database/sql.afterTxInstrumentation
func afterTxInstrumentation(call api.CallContext, tx *sql.Tx, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	callData, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	dbRequest, ok := callData["dbRequest"].(databaseSqlRequest)
	if !ok {
		return
	}
	tx.Endpoint = dbRequest.endpoint
	tx.DriverName = dbRequest.driverName
	tx.DSN = dbRequest.dsn
	instrumentEnd(call, err)
}

//go:linkname beforeConnInstrumentation database/sql.beforeConnInstrumentation
func beforeConnInstrumentation(call api.CallContext, db *sql.DB, ctx context.Context) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if db == nil {
		return
	}
	call.SetData(map[string]string{
		"endpoint": db.Endpoint,
		"driver":   db.DriverName,
		"dsn":      db.DSN,
	})
}

//go:linkname afterConnInstrumentation database/sql.afterConnInstrumentation
func afterConnInstrumentation(call api.CallContext, conn *sql.Conn, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	data, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	endpoint, ok := data["endpoint"]
	if ok {
		conn.Endpoint = endpoint
	}
	driverName, ok := data["driver"]
	if ok {
		conn.DriverName = driverName
	}
	dsn, ok := data["dsn"]
	if ok {
		conn.DSN = dsn
	}
}

//go:linkname beforeConnPingContextInstrumentation database/sql.beforeConnPingContextInstrumentation
func beforeConnPingContextInstrumentation(call api.CallContext, conn *sql.Conn, ctx context.Context) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	instrumentStart(call, ctx, "ping", "ping", conn.Endpoint, conn.DriverName, conn.DSN)
}

//go:linkname afterConnPingContextInstrumentation database/sql.afterConnPingContextInstrumentation
func afterConnPingContextInstrumentation(call api.CallContext, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeConnPrepareContextInstrumentation database/sql.beforeConnPrepareContextInstrumentation
func beforeConnPrepareContextInstrumentation(call api.CallContext, conn *sql.Conn, ctx context.Context, query string) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	call.SetData(map[string]string{
		"endpoint": conn.Endpoint,
		"sql":      query,
		"driver":   conn.DriverName,
		"dsn":      conn.DSN,
	})
}

//go:linkname afterConnPrepareContextInstrumentation database/sql.afterConnPrepareContextInstrumentation
func afterConnPrepareContextInstrumentation(call api.CallContext, stmt *sql.Stmt, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	callDataMap, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	stmt.Data = map[string]string{
		"endpoint": callDataMap["endpoint"],
		"sql":      callDataMap["sql"],
		"driver":   callDataMap["driver"],
	}
	stmt.DSN = callDataMap["dsn"]
}

//go:linkname beforeConnExecContextInstrumentation database/sql.beforeConnExecContextInstrumentation
func beforeConnExecContextInstrumentation(call api.CallContext, conn *sql.Conn, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	instrumentStart(call, ctx, "exec", query, conn.Endpoint, conn.DriverName, conn.DSN, args...)
}

//go:linkname afterConnExecContextInstrumentation database/sql.afterConnExecContextInstrumentation
func afterConnExecContextInstrumentation(call api.CallContext, result sql.Result, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeConnQueryContextInstrumentation database/sql.beforeConnQueryContextInstrumentation
func beforeConnQueryContextInstrumentation(call api.CallContext, conn *sql.Conn, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	instrumentStart(call, ctx, "query", query, conn.Endpoint, conn.DriverName, conn.DSN, args...)
}

//go:linkname afterConnQueryContextInstrumentation database/sql.afterConnQueryContextInstrumentation
func afterConnQueryContextInstrumentation(call api.CallContext, rows *sql.Rows, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeConnTxInstrumentation database/sql.beforeConnTxInstrumentation
func beforeConnTxInstrumentation(call api.CallContext, conn *sql.Conn, ctx context.Context, opts *sql.TxOptions) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if conn == nil {
		return
	}
	instrumentStart(call, ctx, "start", "START TRANSACTION", conn.Endpoint, conn.DriverName, conn.DSN)
}

//go:linkname afterConnTxInstrumentation database/sql.afterConnTxInstrumentation
func afterConnTxInstrumentation(call api.CallContext, tx *sql.Tx, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeTxPrepareContextInstrumentation database/sql.beforeTxPrepareContextInstrumentation
func beforeTxPrepareContextInstrumentation(call api.CallContext, tx *sql.Tx, ctx context.Context, query string) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	call.SetData(map[string]string{
		"endpoint": tx.Endpoint,
		"sql":      query,
		"driver":   tx.DriverName,
		"dsn":      tx.DSN,
	})
}

//go:linkname afterTxPrepareContextInstrumentation database/sql.afterTxPrepareContextInstrumentation
func afterTxPrepareContextInstrumentation(call api.CallContext, stmt *sql.Stmt, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	callDataMap, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	stmt.Data = map[string]string{
		"endpoint": callDataMap["endpoint"],
		"sql":      callDataMap["sql"],
		"driver":   callDataMap["driver"],
	}
	stmt.DSN = callDataMap["dsn"]
}

//go:linkname beforeTxStmtContextInstrumentation database/sql.beforeTxStmtContextInstrumentation
func beforeTxStmtContextInstrumentation(call api.CallContext, tx *sql.Tx, ctx context.Context, stmt *sql.Stmt) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	call.SetData(map[string]string{
		"endpoint": stmt.Data["endpoint"],
		"driver":   stmt.Data["driver"],
		"dsn":      stmt.DSN,
	})
}

//go:linkname afterTxStmtContextInstrumentation database/sql.afterTxStmtContextInstrumentation
func afterTxStmtContextInstrumentation(call api.CallContext, stmt *sql.Stmt) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	data, ok := call.GetData().(map[string]string)
	if !ok {
		return
	}
	stmt.Data = map[string]string{}
	endpoint, ok := data["endpoint"]
	if ok {
		stmt.Data["endpoint"] = endpoint
	}
	driverName, ok := data["driver"]
	if ok {
		stmt.Data["driver"] = driverName
	}
	dsn, ok := data["dsn"]
	if ok {
		stmt.Data["dsn"] = dsn
	}
}

//go:linkname beforeTxExecContextInstrumentation database/sql.beforeTxExecContextInstrumentation
func beforeTxExecContextInstrumentation(call api.CallContext, tx *sql.Tx, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	instrumentStart(call, ctx, "exec", query, tx.Endpoint, tx.DriverName, tx.DSN, args...)
}

//go:linkname afterTxExecContextInstrumentation database/sql.afterTxExecContextInstrumentation
func afterTxExecContextInstrumentation(call api.CallContext, result sql.Result, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeTxQueryContextInstrumentation database/sql.beforeTxQueryContextInstrumentation
func beforeTxQueryContextInstrumentation(call api.CallContext, tx *sql.Tx, ctx context.Context, query string, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	instrumentStart(call, ctx, "query", query, tx.Endpoint, tx.DriverName, tx.DSN, args...)
}

//go:linkname afterTxQueryContextInstrumentation database/sql.afterTxQueryContextInstrumentation
func afterTxQueryContextInstrumentation(call api.CallContext, rows *sql.Rows, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeTxCommitInstrumentation database/sql.beforeTxCommitInstrumentation
func beforeTxCommitInstrumentation(call api.CallContext, tx *sql.Tx) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	instrumentStart(call, context.Background(), "commit", "COMMIT", tx.Endpoint, tx.DriverName, tx.DSN)
}

//go:linkname afterTxCommitInstrumentation database/sql.afterTxCommitInstrumentation
func afterTxCommitInstrumentation(call api.CallContext, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeTxRollbackInstrumentation database/sql.beforeTxRollbackInstrumentation
func beforeTxRollbackInstrumentation(call api.CallContext, tx *sql.Tx) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if tx == nil {
		return
	}
	instrumentStart(call, context.Background(), "rollback", "ROLLBACK", tx.Endpoint, tx.DriverName, tx.DSN)
}

//go:linkname afterTxRollbackInstrumentation database/sql.afterTxRollbackInstrumentation
func afterTxRollbackInstrumentation(call api.CallContext, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeStmtExecContextInstrumentation database/sql.beforeStmtExecContextInstrumentation
func beforeStmtExecContextInstrumentation(call api.CallContext, stmt *sql.Stmt, ctx context.Context, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	sql, endpoint, driverName, dsn := "", "", "", ""
	if stmt.Data != nil {
		sql, endpoint, driverName, dsn = stmt.Data["sql"], stmt.Data["endpoint"], stmt.Data["driver"], stmt.DSN
	}
	instrumentStart(call, ctx, "exec", sql, endpoint, driverName, dsn, args...)
}

//go:linkname afterStmtExecContextInstrumentation database/sql.afterStmtExecContextInstrumentation
func afterStmtExecContextInstrumentation(call api.CallContext, result sql.Result, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}

//go:linkname beforeStmtQueryContextInstrumentation database/sql.beforeStmtQueryContextInstrumentation
func beforeStmtQueryContextInstrumentation(call api.CallContext, stmt *sql.Stmt, ctx context.Context, args ...any) {
	if !dbSqlEnabler.Enable() {
		return
	}
	if stmt == nil {
		return
	}
	sql, endpoint, driverName, dsn := "", "", "", ""
	if stmt.Data != nil {
		sql, endpoint, driverName, dsn = stmt.Data["sql"], stmt.Data["endpoint"], stmt.Data["driver"], stmt.DSN
	}
	instrumentStart(call, ctx, "query", sql, endpoint, driverName, dsn, args...)
}

//go:linkname afterStmtQueryContextInstrumentation database/sql.afterStmtQueryContextInstrumentation
func afterStmtQueryContextInstrumentation(call api.CallContext, rows *sql.Rows, err error) {
	if !dbSqlEnabler.Enable() {
		return
	}
	instrumentEnd(call, err)
}
func instrumentStart(call api.CallContext, ctx context.Context, spanName, query, endpoint, driverName, dsn string, args ...any) {
	req := databaseSqlRequest{
		opType:     calOp(query),
		sql:        query,
		endpoint:   endpoint,
		driverName: driverName,
		dsn:        dsn,
		params:     args,
	}
	newCtx := databaseSqlInstrumenter.Start(ctx, req)
	call.SetData(map[string]interface{}{
		"dbRequest": req,
		"newCtx":    newCtx,
	})
}
func instrumentEnd(call api.CallContext, err error) {
	callData, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}
	dbRequest, ok := callData["dbRequest"].(databaseSqlRequest)
	if !ok {
		return
	}
	newCtx, ok := callData["newCtx"].(context.Context)
	if !ok {
		return
	}
	databaseSqlInstrumenter.End(newCtx, dbRequest, nil, err)
}

func calOp(sql string) string {
	sqls := strings.Split(sql, " ")
	var op string
	if len(sqls) > 0 {
		op = sqls[0]
	}
	return op
}
