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

package utils

import (
	"net/url"
	"strings"
)

type UrlFilter interface {
	FilterUrl(url *url.URL) bool
}

type SpanNameFilter interface {
	FilterSpanName(spanName string) bool
}

type DefaultUrlFilter struct {
}

func (d DefaultUrlFilter) FilterUrl(url *url.URL) bool {
	return false
}

type PathFilter struct {
	paths map[string]bool
}

func NewPathFilter(excludePaths []string) *PathFilter {
	p := &PathFilter{paths: make(map[string]bool)}
	for _, path := range excludePaths {
		p.paths[strings.TrimSpace(path)] = true
	}
	return p
}

func (p *PathFilter) FilterUrl(url *url.URL) bool {
	return p.paths[url.Path]
}
