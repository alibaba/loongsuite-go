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

package http

import "strings"

// httpRouteFromSpanName returns a low-cardinality route template encoded in the
// span name by framework hooks (gin FullPath, mux template, etc.). Raw request
// paths such as "/users/123" must not be inferred from span names.
func httpRouteFromSpanName(spanName string) string {
	if spanName == "" {
		return ""
	}
	path := spanName
	if idx := strings.Index(spanName, " "); idx >= 0 {
		path = spanName[idx+1:]
	}
	if path == "" {
		return ""
	}
	if strings.Contains(path, "{") || strings.Contains(path, "*") {
		return path
	}
	if strings.Contains(path, ":") {
		return path
	}
	return ""
}

// netHttpRouteFromPattern normalizes net/http Request.Pattern (Go 1.22+).
func netHttpRouteFromPattern(pattern string) string {
	if pattern == "" {
		return ""
	}
	if strings.HasPrefix(pattern, "/") {
		return pattern
	}
	if _, rest, ok := strings.Cut(pattern, " "); ok {
		if idx := strings.Index(rest, "/"); idx >= 0 {
			return rest[idx:]
		}
	}
	return ""
}
