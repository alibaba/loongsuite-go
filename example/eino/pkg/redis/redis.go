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

package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	RedisPrefix = "eino:doc:"
	IndexName   = "vector_index"

	ContentField  = "content"
	MetadataField = "metadata"
	VectorField   = "content_vector"
	DistanceField = "distance"
)

var initOnce sync.Once

func Init() error {
	var err error
	initOnce.Do(func() {
		err = InitRedisIndex(context.Background(), &Config{
			RedisAddr: "localhost:6379",
			Dimension: 4096,
		})
	})
	return err
}

type Config struct {
	RedisAddr string
	Dimension int
}

func InitRedisIndex(ctx context.Context, config *Config) (err error) {
	if config.Dimension <= 0 {
		return fmt.Errorf("dimension must be positive")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Protocol: 2,
	})

	defer func() {
		if err != nil {
			client.Close()
		}
	}()

	if err = client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	indexName := fmt.Sprintf("%s%s", RedisPrefix, IndexName)

	// 检查是否存在索引
	exists, err := client.Do(ctx, "FT.INFO", indexName).Result()
	if err != nil {
		if !strings.Contains(err.Error(), "Unknown index name") {
			return fmt.Errorf("failed to check if index exists: %w", err)
		}
		err = nil
	} else if exists != nil {
		return nil
	}

	// Create new index
	createIndexArgs := []interface{}{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", RedisPrefix,
		"SCHEMA",
		ContentField, "TEXT",
		MetadataField, "TEXT",
		VectorField, "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", config.Dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	if err = client.Do(ctx, createIndexArgs...).Err(); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// 验证索引是否创建成功
	if _, err = client.Do(ctx, "FT.INFO", indexName).Result(); err != nil {
		return fmt.Errorf("failed to verify index creation: %w", err)
	}

	return nil
}
