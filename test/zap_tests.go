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
	"strings"
	"testing"
)

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("zap-test", "zap", "", "", "1.21", "", TestZap),
		NewGeneralTestCase("zap-test-with-field", "zap", "", "", "1.21", "", TestZapWithField),
	)
}

func TestZap(t *testing.T, env ...string) {
	UseApp("zap")
	RunGoBuild(t, "go", "build", "test_zap.go", "http_server.go")
	_, stderr := RunApp(t, "test_zap", env...)
	reader := strings.NewReader(stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[test debugging]") {
			continue
		}
		ExpectContains(t, line, "trace_id")
		ExpectContains(t, line, "span_id")
	}
}
func TestZapWithField(t *testing.T, env ...string) {
	UseApp("zap")
	RunGoBuild(t, "go", "build", "test_zap_with_field.go")
	_, stderr := RunApp(t, "test_zap_with_field", env...)
	reader := strings.NewReader(stderr)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[test debugging]") {
			continue
		}
		ExpectContains(t, line, "trace_id")
		ExpectContains(t, line, "span_id")
	}
}
