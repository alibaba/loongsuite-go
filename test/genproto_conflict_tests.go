// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

// TestGenprotoConflict tests the fix for issue #627
// https://github.com/alibaba/loongsuite-go-agent/issues/627
//
// This test verifies that the otel tool can handle projects with dependencies
// that use the old monolithic google.golang.org/genproto module (like aliyun-odps-go-sdk)
// alongside OpenTelemetry dependencies that use the newer split modules
// (google.golang.org/genproto/googleapis/api and google.golang.org/genproto/googleapis/rpc).
//
// Before the fix, building such projects would fail with "ambiguous import" errors.
// After the fix, the otel tool explicitly pins the split module versions to resolve conflicts.
func TestGenprotoConflict(t *testing.T) {
	const AppName = "genproto-conflict"
	UseApp(AppName)

	// Build should succeed without "ambiguous import" errors
	RunGoBuild(t, "go", "build", "-o", "demo", "main.go")

	// Verify that no ambiguous import errors occurred during build
	debugLog := ReadLog(t)
	ExpectNotContains(t, debugLog, "ambiguous import")
	ExpectNotContains(t, debugLog, "found package google.golang.org/genproto")

	t.Log("Successfully built project with genproto dependency conflict resolution (issue #627)")
}
