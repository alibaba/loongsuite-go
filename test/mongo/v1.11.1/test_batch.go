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

	// Creates write models that specify replace and update operations
	models := []mongo.WriteModel{
		mongo.NewReplaceOneModel().SetFilter(bson.D{{"name", "Cafe Tomato"}}).
			SetReplacement(Restaurant{Name: "Cafe Zucchini", Cuisine: "French"}),
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"name", "Cafe Zucchini"}}).
			SetUpdate(bson.D{{"$set", bson.D{{"name", "Zucchini Land"}}}}),
	}

	// Specifies that the bulk write is ordered
	opts := options.BulkWrite().SetOrdered(true)

	// Runs a bulk write operation for the specified write operations
	_, err = coll.BulkWrite(context.TODO(), models, opts)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "update", "mongodb", "127.0.0.1", "update", "update", "", nil)
	}, 1)
}
