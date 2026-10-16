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

// RouteFromSpanName returns a low-cardinality route template encoded in the
// span name by framework hooks (gin FullPath, mux template, etc.). Raw request
// paths such as "/users/123" must not be inferred from span names.
func RouteFromSpanName(spanName string) string {
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
	if strings.Contains(path, "{") && strings.Contains(path, "}") {
		return path
	}
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
			return path
		}
	}
	return ""
}

// NetHttpRouteFromPattern normalizes net/http Request.Pattern (Go 1.22+).
func NetHttpRouteFromPattern(pattern string) string {
	if pattern == "" {
		return ""
	}
	if _, rest, ok := strings.Cut(pattern, " "); ok {
		pattern = rest
	}
	if idx := strings.IndexByte(pattern, '/'); idx >= 0 {
		return pattern[idx:]
	}
	return ""
}
