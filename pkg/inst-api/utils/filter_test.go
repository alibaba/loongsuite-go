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
	"testing"
)

func TestDefaultUrlFilter(t *testing.T) {
	filter := DefaultUrlFilter{}
	testCases := []struct {
		input    *url.URL
		expected bool
	}{
		{
			input:    &url.URL{Scheme: "http", Host: "example.com"},
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input.String(), func(t *testing.T) {
			result := filter.FilterUrl(tc.input)
			if result != tc.expected {
				t.Errorf("FilterUrl(%v) = %v; expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestPathFilter(t *testing.T) {
	filter := NewPathFilter([]string{"/ping", "/health", " /metrics "})
	testCases := []struct {
		name     string
		input    *url.URL
		expected bool
	}{
		{
			name:     "excluded path /ping",
			input:    &url.URL{Scheme: "http", Host: "example.com", Path: "/ping"},
			expected: true,
		},
		{
			name:     "excluded path /health",
			input:    &url.URL{Scheme: "http", Host: "example.com", Path: "/health"},
			expected: true,
		},
		{
			name:     "excluded path /metrics with trimmed spaces",
			input:    &url.URL{Scheme: "http", Host: "example.com", Path: "/metrics"},
			expected: true,
		},
		{
			name:     "non-excluded path /api/users",
			input:    &url.URL{Scheme: "http", Host: "example.com", Path: "/api/users"},
			expected: false,
		},
		{
			name:     "non-excluded root path",
			input:    &url.URL{Scheme: "http", Host: "example.com", Path: "/"},
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter.FilterUrl(tc.input)
			if result != tc.expected {
				t.Errorf("FilterUrl(%v) = %v; expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestPathFilterEmpty(t *testing.T) {
	filter := NewPathFilter([]string{})
	u := &url.URL{Scheme: "http", Host: "example.com", Path: "/ping"}
	if filter.FilterUrl(u) {
		t.Errorf("FilterUrl(%v) = true; expected false for empty filter", u)
	}
}
