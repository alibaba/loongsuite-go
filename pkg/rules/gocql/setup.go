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

package gocql

import (
	"context"
	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/gocql/gocql"
	"os"
	"strings"
	_ "unsafe"
)

type gocqlInnerEnabler struct {
	enabled bool
}

func (g gocqlInnerEnabler) Enable() bool {
	return g.enabled
}

var gocqlEnabler = gocqlInnerEnabler{
	enabled: os.Getenv("OTEL_INSTRUMENTATION_GOCQL_ENABLED") != "false",
}

var gocqlInstrumenter = BuildGocqlInstrumenter()

//go:linkname beforeCreateSession github.com/gocql/gocql.beforeCreateSession
func beforeCreateSession(_ api.CallContext, clusterCfg *gocql.ClusterConfig) {
	if !gocqlEnabler.Enable() {
		return
	}
	if clusterCfg == nil {
		return
	}
	otelObsvr := newOtelObserver(clusterCfg.QueryObserver, clusterCfg.BatchObserver)
	clusterCfg.QueryObserver = otelObsvr
	clusterCfg.BatchObserver = otelObsvr
	// try to fill user
	if clusterCfg.Authenticator == nil {
		return
	}
	passwordAuthenticator, ok := clusterCfg.Authenticator.(gocql.PasswordAuthenticator)
	if !ok {
		return
	}
	otelObsvr.user = passwordAuthenticator.Username
}

type otelObserver struct {
	queryObserver gocql.QueryObserver
	batchObserver gocql.BatchObserver
	user          string
}

func newOtelObserver(queryObserver gocql.QueryObserver, batchObserver gocql.BatchObserver) *otelObserver {
	return &otelObserver{
		queryObserver: queryObserver,
		batchObserver: batchObserver,
	}
}

func (o *otelObserver) ObserveQuery(ctx context.Context, query gocql.ObservedQuery) {
	request := gocqlRequest{
		Statement: query.Statement,
		DbName:    query.Keyspace,
		Addr:      query.Host.HostnameAndPort(),
		Op:        extractOpType(query.Statement),
		User:      o.user,
		BatchSize: 1,
	}
	gocqlInstrumenter.StartAndEnd(ctx, request, nil, query.Err, query.Start, query.End)
	if o.queryObserver != nil {
		o.queryObserver.ObserveQuery(ctx, query)
	}
}

func (o *otelObserver) ObserveBatch(ctx context.Context, batch gocql.ObservedBatch) {
	request := gocqlRequest{
		Statement: strings.Join(batch.Statements, ", "),
		DbName:    batch.Keyspace,
		Addr:      batch.Host.HostnameAndPort(),
		Op:        "BATCH",
		User:      o.user,
		BatchSize: len(batch.Statements),
	}
	gocqlInstrumenter.StartAndEnd(ctx, request, nil, batch.Err, batch.Start, batch.End)
	if o.batchObserver != nil {
		o.batchObserver.ObserveBatch(ctx, batch)
	}
}
