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

package databasesql

import (
	"github.com/cespare/xxhash/v2"
	lru "github.com/hashicorp/golang-lru/v2"
)

type SQLMeta struct {
	stmt       string
	operation  string
	collection string
	params     []any
}

type SQLMetaCache struct {
	cache *lru.Cache[uint64, SQLMeta]
}

func NewSQLMetaCache(size int) (*SQLMetaCache, error) {
	cache, err := lru.New[uint64, SQLMeta](size)
	if err != nil {
		return nil, err
	}

	return &SQLMetaCache{
		cache: cache,
	}, nil
}

func computeHash(sql string) uint64 {
	return xxhash.Sum64String(sql)
}

func (c *SQLMetaCache) Get(key string) (SQLMeta, bool) {
	hash := computeHash(key)
	return c.cache.Get(hash)
}

func (c *SQLMetaCache) Add(key string, value SQLMeta) bool {
	hash := computeHash(key)
	return c.cache.Add(hash, value)
}

func (c *SQLMetaCache) Remove(key string) bool {
	hash := computeHash(key)
	return c.cache.Remove(hash)
}

func (c *SQLMetaCache) Contains(key string) bool {
	hash := computeHash(key)
	return c.cache.Contains(hash)
}

func (c *SQLMetaCache) Len() int {
	return c.cache.Len()
}
