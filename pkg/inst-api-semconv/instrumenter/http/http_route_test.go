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

import "testing"

func TestHttpRouteFromSpanName(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"GET /users/123", ""},
		{"/user/:name", "/user/:name"},
		{"/test/{key}", "/test/{key}"},
		{"/hertz/:version/*action", "/hertz/:version/*action"},
		{"GET /a", ""},
	} {
		if got := httpRouteFromSpanName(tc.name); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestNetHttpRouteFromPattern(t *testing.T) {
	if got := netHttpRouteFromPattern("GET /users/{id}"); got != "/users/{id}" {
		t.Fatalf("got %q", got)
	}
	if got := netHttpRouteFromPattern("/health"); got != "/health" {
		t.Fatalf("got %q", got)
	}
	if got := netHttpRouteFromPattern(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
