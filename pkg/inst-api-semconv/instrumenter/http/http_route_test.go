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

func TestRouteFromSpanName(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"GET /users/123", ""},
		{"/user/:name", "/user/:name"},
		{"/test/{key}", "/test/{key}"},
		{"/hertz/:version/*action", "/hertz/:version/*action"},
		{"GET /a", ""},
		{"GET /time/12:30", ""},
		{"GET /files/*path", "/files/*path"},
		{"GET /literal*value", ""},
	} {
		if got := RouteFromSpanName(tc.name); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestNetHttpRouteFromPattern(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		want    string
	}{
		{"GET /users/{id}", "/users/{id}"},
		{"/health", "/health"},
		{"example.com/x", "/x"},
		{"GET example.com/x/{id}", "/x/{id}"},
		{"/files/{path...}", "/files/{path...}"},
		{"/{$}", "/{$}"},
		{"", ""},
	} {
		if got := NetHttpRouteFromPattern(tc.pattern); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.pattern, got, tc.want)
		}
	}
}
