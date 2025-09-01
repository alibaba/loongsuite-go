// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package test

import (
	"path/filepath"
	"regexp"
	"testing"
)

const ErrorsAppName = "errorstest"

func TestRunErrors(t *testing.T) {
	UseApp(ErrorsAppName)
	RunSet(t, UseTestRules("test_error.json"))
	RunGoBuild(t, "go", "build")
	stdout, stderr := RunApp(t, ErrorsAppName)
	ExpectContains(t, stdout, "wow")
	ExpectContains(t, stdout, "old:wow")
	ExpectContains(t, stdout, "ptr<nil>")
	ExpectNotContains(t, stdout, "val1024")
	ExpectContains(t, stdout, "val1298") // 0x512
	ExpectContains(t, stdout, "7632")
	ExpectContains(t, stdout, "4008208820")
	ExpectContains(t, stdout, "118888")
	ExpectContains(t, stdout, "0.001")
	ExpectContains(t, stderr, "2024 shanghai")
	ExpectContains(t, stdout, "2033 hangzhou")
	ExpectNotContains(t, stderr, "failed to exec")
	ExpectNotContains(t, stderr, "baddep")
	ExpectContains(t, stderr, "gooddep")
	text := ReadInstrumentLog(t, filepath.Join("auxiliary", "helper.go"))
	re := regexp.MustCompile(".*OtelOnEnterTrampoline_TestSkip.*")
	matches := re.FindAllString(text, -1)
	if len(matches) < 1 {
		t.Fatalf("expecting at least one match")
	}
	re = regexp.MustCompile(".*OtelOnEnterTrampoline_p1.*")
	matches = re.FindAllString(text, -1)
	if len(matches) != 4 {
		t.Fatalf("expecting 4 matches")
	}
	re = regexp.MustCompile(".*OtelOnExitTrampoline_p2.*")
	matches = re.FindAllString(text, -1)
	if len(matches) != 4 {
		t.Fatalf("expecting 4 matches")
	}

	// Test for generic hook
	maxFunc := 7
	re = regexp.MustCompile(".*xian.*")
	matches = re.FindAllString(stderr, -1)
	if len(matches) != 4 { // f1 + f2 + f4 + init
		t.Fatalf("expecting 4 matches")
	}
	re = regexp.MustCompile(".*shanxi.*") // f1 + f2
	matches = re.FindAllString(stderr, -1)
	if len(matches) != 2 {
		t.Fatalf("expecting 2 matches")
	}
	re = regexp.MustCompile(".*zhejiang.*") // match all funcs(including init)
	matches = re.FindAllString(stderr, -1)
	if len(matches) != maxFunc {
		t.Fatalf("expecting full matches")
	}
	re = regexp.MustCompile(".*beijing.*") // f3 + f5
	matches = re.FindAllString(stderr, -1)
	if len(matches) != 2 {
		t.Fatalf("expecting 2 matches")
	}
	re = regexp.MustCompile(".*entering.*") // match all funcs(including init)
	matches = re.FindAllString(stderr, -1)
	if len(matches) != maxFunc {
		t.Fatalf("expecting full matches")
	}
	re = regexp.MustCompile(".*within.*") // match all funcs(including init)
	matches = re.FindAllString(stderr, -1)
	if len(matches) != maxFunc {
		t.Fatalf("expecting full matches")
	}
}
