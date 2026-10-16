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

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSetServerRoutePattern(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pattern     string
		path        string
		spanName    string
		wantPattern string
	}{
		{
			name:        "framework template overrides ServeMux catch-all",
			pattern:     "/",
			path:        "/users/123",
			spanName:    "/users/{user-id}",
			wantPattern: "/users/{user-id}",
		},
		{
			name:        "framework static route overrides ServeMux catch-all",
			pattern:     "/",
			path:        "/health",
			spanName:    "/health",
			wantPattern: "/health",
		},
		{
			name:        "ServeMux pattern is retained",
			pattern:     "GET /users/{id}",
			path:        "/users/123",
			spanName:    "GET",
			wantPattern: "GET /users/{id}",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := &netHttpRequest{}
			rawRequest := &http.Request{
				Pattern: tc.pattern,
				URL:     &url.URL{Path: tc.path},
			}

			setServerRoutePattern(request, rawRequest, tc.spanName)

			if request.routePattern != tc.wantPattern {
				t.Fatalf("route pattern = %q, want %q", request.routePattern, tc.wantPattern)
			}
		})
	}
}
