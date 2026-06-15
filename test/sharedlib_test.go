// Copyright (c) 2025 Alibaba Group Holding Ltd.
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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestLoadSharedLib(t *testing.T) {
	const AppName = "sharedlib"
	UseApp(AppName + "/so")
	tempDir := filepath.Join(os.TempDir(), uuid.New().String())
	RunSet(t, "-pkg="+tempDir)
	RunGoBuild(t, "go", "build", "-buildmode=plugin", "-o=plugin.so", "plugin.go")
	UseApp(AppName)
	RunSet(t, "-pkg="+tempDir)
	RunGoBuild(t, "go", "build")
	stdout, _ := RunApp(t, AppName)
	ExpectContains(t, stdout, "plugin init")
}
