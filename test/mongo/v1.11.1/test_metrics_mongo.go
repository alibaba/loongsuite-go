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
	"log"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// User model.
type User struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	Age  int                `bson:"age"`
}

func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn))
	if err != nil {
		panic(fmt.Sprintf("connect mongodb error %v \n", err))
	}
	ctx := context.Background()
	err = TestCreateCollection(ctx, client)
	if err != nil {
		log.Printf("failed to create collection: %v", err)
	}
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		// TODO : verifier
	})
}

func TestCreateCollection(ctx context.Context, client *mongo.Client) error {
	return client.Database(db).CreateCollection(ctx, "users")
}
