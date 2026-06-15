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

package clickhousev2

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"os"
	"strings"
	"time"
	_ "unsafe"
)

type clickhouseInnerEnabled struct {
	enabled bool
}

func (c *clickhouseInnerEnabled) Enable() bool {
	return c.enabled
}

var innerEnabled = clickhouseInnerEnabled{enabled: os.Getenv("OTEL_INSTRUMENTATION_CLICKHOUSE_V2_ENABLED") != "false"}

var clickhouseInstrumenter = BuildClickhouseInstrumenter()

func NewOtelCon(con driver.Conn, opts *clickhouse.Options) *OtelCon {
	return &OtelCon{con: con, opts: opts}
}

type OtelCon struct {
	con  driver.Conn
	opts *clickhouse.Options
}

func (oc *OtelCon) Contributors() []string {
	return oc.con.Contributors()
}

func (oc *OtelCon) ServerVersion() (*driver.ServerVersion, error) {
	startTime := time.Now()
	sv, err := oc.con.ServerVersion()
	request := clickhouseRequest{
		Statement: "SERVER_VERSION",
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "SERVER_VERSION",
		BatchSize: 1,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return sv, err
}

func (oc *OtelCon) Select(ctx context.Context, dest any, query string, args ...any) error {
	startTime := time.Now()
	err := oc.con.Select(ctx, dest, query, args...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "SELECT",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return err
}

func (oc *OtelCon) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	startTime := time.Now()
	rows, err := oc.con.Query(ctx, query, args...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "QUERY",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return rows, err
}

func (oc *OtelCon) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	startTime := time.Now()
	row := oc.con.QueryRow(ctx, query, args...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "QUERY_ROW",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, row.Err(), startTime, time.Now())
	return row
}

func (oc *OtelCon) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	startTime := time.Now()
	batch, err := oc.con.PrepareBatch(ctx, query, opts...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "PREPARE_BATCH",
		BatchSize: 1,
		Params:    nil,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return batch, err
}

func (oc *OtelCon) Exec(ctx context.Context, query string, args ...any) error {
	startTime := time.Now()
	err := oc.con.Exec(ctx, query, args...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "EXEC",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return err
}

func (oc *OtelCon) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	startTime := time.Now()
	err := oc.con.AsyncInsert(ctx, query, wait, args...)
	request := clickhouseRequest{
		Statement: query,
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "ASYNC_INSERT",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return err
}

func (oc *OtelCon) Ping(ctx context.Context) error {
	startTime := time.Now()
	err := oc.con.Ping(ctx)
	request := clickhouseRequest{
		Statement: "PING",
		DbName:    oc.opts.Auth.Database,
		User:      oc.opts.Auth.Username,
		Addr:      strings.Join(oc.opts.Addr, ","),
		Op:        "PING",
		BatchSize: 1,
		Params:    nil,
	}
	clickhouseInstrumenter.StartAndEnd(context.Background(), request, nil, err, startTime, time.Now())
	return err
}

func (oc *OtelCon) Stats() driver.Stats {
	return oc.con.Stats()
}

func (oc *OtelCon) Close() error {
	return oc.con.Close()
}

//go:linkname beforeOpen github.com/ClickHouse/clickhouse-go/v2.beforeOpen
func beforeOpen(ctx api.CallContext, options *clickhouse.Options) {
	if !innerEnabled.Enable() {
		return
	}
	ctx.SetData(options)
}

//go:linkname afterOpen github.com/ClickHouse/clickhouse-go/v2.afterOpen
func afterOpen(ctx api.CallContext, con driver.Conn, err error) {
	if err != nil {
		return
	}
	if !innerEnabled.Enable() {
		return
	}
	ctx.SetReturnVal(0, NewOtelCon(con, ctx.GetData().(*clickhouse.Options)))
}
