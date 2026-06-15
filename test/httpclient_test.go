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
	"testing"
)

const HttpclientAppName = "httpclient"

func TestRunHttpclient(t *testing.T) {
	UseApp(HttpclientAppName)

	RunSet(t, UseTestRules("test_nethttp.json"), "-verbose")
	RunGoBuild(t, "go", "build")
	_, stderr := RunApp(t, HttpclientAppName)
	ExpectContains(t, stderr, "Client.Do()")                // println writes to stderr
	ExpectContains(t, stderr, "failed to exec onExit hook") // intentional panic
	ExpectContains(t, stderr, "NewRequest()")
	ExpectContains(t, stderr, "NewRequest1()")
	ExpectContains(t, stderr, "NewRequestWithContext()")
	ExpectContains(t, stderr, "MaxBytesError()")
	ExpectContains(t, stderr, "debug.Stack()") // during recover()
	ExpectContains(t, stderr, "4008208820")
	ExpectContains(t, stderr, "Prince of Qin Smashing the Battle line")
	ExpectContains(t, stderr, "IsPrint")
}
