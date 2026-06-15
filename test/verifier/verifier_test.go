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

package verifier

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

func TestNoSqlAttributesPass(t *testing.T) {
	VerifyDbAttributes(tracetest.SpanStub{SpanKind: trace.SpanKindClient, Name: "name", Attributes: []attribute.KeyValue{
		{Key: semconv.DBNameKey, Value: attribute.StringValue("dbname")},
		{Key: semconv.DBSystemKey, Value: attribute.StringValue("system")},
		{Key: semconv.DBUserKey, Value: attribute.StringValue("user")},
		{Key: semconv.DBConnectionStringKey, Value: attribute.StringValue("connString")},
		{Key: semconv.DBStatementKey, Value: attribute.StringValue("statement")},
		{Key: semconv.DBOperationKey, Value: attribute.StringValue("operation")},
		{Key: semconv.DBCollectionNameKey, Value: attribute.StringValue("collection")},
		{Key: attribute.Key("db.query.parameter"), Value: attribute.StringSliceValue([]string{"param0"})},
	}}, "name", "system", "connString", "statement", "operation", "collection", []any{"param0"})
}

func TestNoSqlAttributesFail(t *testing.T) {
	defer func() {
		pass := false
		if r := recover(); r != nil {
			pass = true
		}
		if !pass {
			t.Fatal("Should be recovered from panic")
		}
	}()
	VerifyDbAttributes(tracetest.SpanStub{SpanKind: trace.SpanKindClient, Name: "name", Attributes: []attribute.KeyValue{
		{Key: semconv.DBNameKey, Value: attribute.StringValue("dbname")},
		{Key: semconv.DBSystemKey, Value: attribute.StringValue("system")},
		{Key: semconv.DBUserKey, Value: attribute.StringValue("user")},
		{Key: semconv.DBConnectionStringKey, Value: attribute.StringValue("connString")},
		{Key: semconv.DBStatementKey, Value: attribute.StringValue("wrong statement")},
		{Key: semconv.DBOperationKey, Value: attribute.StringValue("operation")},
		{Key: semconv.DBCollectionNameKey, Value: attribute.StringValue("collection")},
		{Key: attribute.Key("db.query.parameter"), Value: attribute.StringSliceValue([]string{"param0"})},
	}}, "name", "system", "connString", "statement", "operation", "collection", []any{"param0"})
}
