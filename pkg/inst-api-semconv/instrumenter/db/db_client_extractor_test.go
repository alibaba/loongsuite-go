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

package db

import (
	"context"
	"log"
	"testing"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type testRequest struct {
	Name      string
	Operation string
	Target    string
}

type testResponse struct {
}

type mongoAttrsGetter struct {
}

func (m mongoAttrsGetter) GetSystem(request testRequest) string {
	if request.Name != "" {
		return request.Name
	}
	return ""
}

func (m mongoAttrsGetter) GetServerAddress(request testRequest) string {
	return "test"
}

func (m mongoAttrsGetter) GetStatement(request testRequest) string {
	return "test"
}

func (m mongoAttrsGetter) GetOperation(request testRequest) string {
	if request.Operation != "" {
		return request.Operation
	}
	return ""
}
func (m mongoAttrsGetter) GetDbNamespace(request testRequest) string {
	return ""
}
func (m mongoAttrsGetter) GetBatchSize(request testRequest) int {
	return 0
}

func (m mongoAttrsGetter) GetParameters(request testRequest) []any {
	return nil
}

func (m mongoAttrsGetter) GetCollection(request testRequest) string {
	return request.Target
}

func TestGetSpanKey(t *testing.T) {
	dbExtractor := &DbClientAttrsExtractor[testRequest, any, mongoAttrsGetter]{}
	if dbExtractor.GetSpanKey() != utils.DB_CLIENT_KEY {
		t.Fatalf("Should have returned DB_CLIENT_KEY")
	}
}

func TestDbCommonGetSpanKey(t *testing.T) {
	dbExtractor := &DbClientCommonAttrsExtractor[testRequest, any, mongoAttrsGetter]{}
	if dbExtractor.GetSpanKey() != utils.DB_CLIENT_KEY {
		t.Fatalf("Should have returned DB_CLIENT_KEY")
	}
}

func TestDbClientExtractorStart(t *testing.T) {
	dbExtractor := DbClientAttrsExtractor[testRequest, testResponse, mongoAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = dbExtractor.OnStart(attrs, parentContext, testRequest{Name: "test"})
	if len(attrs) != 0 {
		log.Fatal("attrs should be empty")
	}
}

func TestDbClientExtractorEnd(t *testing.T) {
	dbExtractor := DbClientAttrsExtractor[testRequest, testResponse, mongoAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = dbExtractor.OnEnd(attrs, parentContext, testRequest{Name: "test"}, testResponse{}, nil)
	if attrs[0].Key != semconv.DBSystemNameKey || attrs[0].Value.AsString() != "test" {
		t.Fatalf("db system should be test")
	}
	if attrs[1].Key != semconv.DBQueryTextKey || attrs[1].Value.AsString() != "test" {
		t.Fatalf("db user should be test")
	}
	if attrs[2].Key != semconv.DBOperationNameKey || attrs[2].Value.AsString() != "" {
		t.Fatalf("db connection key should be empty")
	}
	if attrs[3].Key != semconv.ServerAddressKey || attrs[3].Value.AsString() != "test" {
		t.Fatalf("db statement key should be test")
	}
}

func TestDbClientExtractorWithFilter(t *testing.T) {
	dbExtractor := DbClientAttrsExtractor[testRequest, testResponse, mongoAttrsGetter]{}
	dbExtractor.Base.AttributesFilter = func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return []attribute.KeyValue{{
			Key:   "test",
			Value: attribute.StringValue("test"),
		}}
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = dbExtractor.OnStart(attrs, parentContext, testRequest{Name: "test"})
	if attrs[0].Key != "test" || attrs[0].Value.AsString() != "test" {
		panic("attribute should be test")
	}
	attrs, _ = dbExtractor.OnEnd(attrs, parentContext, testRequest{Name: "test"}, testResponse{}, nil)
	if attrs[0].Key != "test" || attrs[0].Value.AsString() != "test" {
		panic("attribute should be test")
	}
}
