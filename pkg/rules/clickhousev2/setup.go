package clickhousev2

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"os"
	"strings"
)

type clickhouseInnerEnabled struct {
	enabled bool
}

func (c *clickhouseInnerEnabled) Enable() bool {
	return c.enabled
}

var innerEnabled = clickhouseInnerEnabled{enabled: os.Getenv("OTEL_INSTRUMENTATION_CLICKHOUSE_V2_ENABLED") != "false"}

var clickhouseInstrumenter = BuildClickhouseInstrumenter()

//go:linkname beforeServerVersion github.com/ClickHouse/clickhouse-go/v2.beforeServerVersion
func beforeServerVersion(ctx api.CallContext, con driver.Conn) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: "SERVER_VERSION",
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "SERVER_VERSION",
		BatchSize: 1,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterServerVersion github.com/ClickHouse/clickhouse-go/v2.afterServerVersion
func afterServerVersion(ctx api.CallContext, con driver.Conn, _ *clickhouse.ServerVersion, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeSelect github.com/ClickHouse/clickhouse-go/v2.beforeSelect
func beforeSelect(ctx api.CallContext, con driver.Conn, _ context.Context, _ any, query string, args ...any) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "SELECT",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQuery github.com/ClickHouse/clickhouse-go/v2.afterSelect
func afterSelect(ctx api.CallContext, _ driver.Conn, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeQuery github.com/ClickHouse/clickhouse-go/v2.beforeQuery
func beforeQuery(ctx api.CallContext, con driver.Conn, _ context.Context, query string, args ...any) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "QUERY",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQuery github.com/ClickHouse/clickhouse-go/v2.afterQuery
func afterQuery(ctx api.CallContext, _ driver.Conn, _ driver.Rows, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeQueryRow github.com/ClickHouse/clickhouse-go/v2.beforeQueryRow
func beforeQueryRow(ctx api.CallContext, con driver.Conn, _ context.Context, query string, args ...any) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "QUERY_ROW",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQueryRow github.com/ClickHouse/clickhouse-go/v2.afterQueryRow
func afterQueryRow(ctx api.CallContext, _ driver.Conn, row driver.Row) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, row.Err())
}

//go:linkname beforePrepareBatch github.com/ClickHouse/clickhouse-go/v2.beforePrepareBatch
func beforePrepareBatch(ctx api.CallContext, con driver.Conn, _ context.Context, query string, _ ...driver.PrepareBatchOption) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "PREPARE_BATCH",
		BatchSize: 1,
		Params:    nil,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterPrepareBatch github.com/ClickHouse/clickhouse-go/v2.afterPrepareBatch
func afterPrepareBatch(ctx api.CallContext, _ driver.Conn, _ driver.Batch, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeExec github.com/ClickHouse/clickhouse-go/v2.beforeExec
func beforeExec(ctx api.CallContext, con driver.Conn, _ context.Context, query string, args ...any) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "EXEC",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterExec github.com/ClickHouse/clickhouse-go/v2.afterExec
func afterExec(ctx api.CallContext, _ driver.Conn, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeAsyncInsert github.com/ClickHouse/clickhouse-go/v2.beforeAsyncInsert
func beforeAsyncInsert(ctx api.CallContext, con driver.Conn, _ context.Context, query string, _ bool, args ...any) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: query,
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "ASYNC_INSERT",
		BatchSize: 1,
		Params:    args,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterAsyncInsert github.com/ClickHouse/clickhouse-go/v2.afterAsyncInsert
func afterAsyncInsert(ctx api.CallContext, _ driver.Conn, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforePing github.com/ClickHouse/clickhouse-go/v2.beforePing
func beforePing(ctx api.CallContext, con driver.Conn, _ context.Context) {
	if !innerEnabled.Enable() {
		return
	}
	ck, ok := con.(*clickhouse.clickhouse)
	if !ok {
		return
	}
	request := clickhouseRequest{
		Statement: "PING",
		DbName:    ck.opt.Auth.Database,
		User:      ck.opt.Auth.Username,
		Addr:      strings.Join(ck.opt.Addr, ","),
		Op:        "PING",
		BatchSize: 1,
		Params:    nil,
	}
	clickhouseInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterPing github.com/ClickHouse/clickhouse-go/v2.afterPing
func afterPing(ctx api.CallContext, _ driver.Conn, err error) {
	if !innerEnabled.Enable() {
		return
	}
	request := ctx.GetData().(clickhouseRequest)
	clickhouseInstrumenter.End(context.Background(), request, nil, err)
}
