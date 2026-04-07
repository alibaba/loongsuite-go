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

package test

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("golog-test", "golog", "", "", "1.18", "", TestGoLog),
	)
}

func TestGoLog(t *testing.T, env ...string) {
	UseApp("golog")
	RunGoBuild(t, "go", "build", "test_glog.go")
	_, stderr := RunApp(t, "test_glog", env...)
	reader := strings.NewReader(stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[test debugging]") {
			continue
		}
		// Check that trace_id and span_id appear in the log line
		ExpectContains(t, line, "trace_id")
		ExpectContains(t, line, "span_id")
		// For JSON lines (slog JSON handler), verify trace_id and span_id are
		// top-level structured fields, not embedded in the msg string.
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			var fields map[string]interface{}
			if err := json.Unmarshal([]byte(line), &fields); err == nil {
				if _, ok := fields["trace_id"]; !ok {
					t.Errorf("expected trace_id to be a structured JSON field in: %s", line)
				}
				if _, ok := fields["span_id"]; !ok {
					t.Errorf("expected span_id to be a structured JSON field in: %s", line)
				}
				if msg, ok := fields["msg"].(string); ok {
					if strings.Contains(msg, "trace_id") {
						t.Errorf("trace_id should not be embedded in msg, got msg: %s", msg)
					}
					if strings.Contains(msg, "span_id") {
						t.Errorf("span_id should not be embedded in msg, got msg: %s", msg)
					}
				}
			}
		}
	}
}
