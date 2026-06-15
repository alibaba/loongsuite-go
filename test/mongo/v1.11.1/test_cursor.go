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

package main

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn))
	if err != nil {
		panic(fmt.Sprintf("connect mongodb error %v \n", err))
	}
	coll := client.Database("sample_restaurants").Collection("restaurants")

	// Creates a query filter to match documents in which the "cuisine"
	// is "Italian"
	filter := bson.D{{"cuisine", "Italian"}}

	// Retrieves documents that match the query filter
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	// Unpacks the cursor into a slice
	var results []Restaurant
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "find", "mongodb", "127.0.0.1", "find", "find", "", nil)
	}, 1)
}
