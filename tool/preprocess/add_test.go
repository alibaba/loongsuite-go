// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package preprocess

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func findReplaceVersion(t *testing.T, gomod, path string) (string, bool) {
	t.Helper()
	modfile, err := parseGoMod(gomod)
	if err != nil {
		t.Fatalf("failed to parse go.mod: %v", err)
	}
	for _, replace := range modfile.Replace {
		if replace.Old.Path == path {
			return replace.New.Version, true
		}
	}
	return "", false
}

func TestPinConflictingHookDependenciesUsesOriginalUserVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findReplaceVersion(t, gomod, "example.com/dep")
	if !ok {
		t.Fatalf("expected replace for example.com/dep")
	}
	if version != "v1.0.0" {
		t.Fatalf("replace version = %q, want %q", version, "v1.0.0")
	}
}

func TestPinConflictingHookDependenciesIgnoresIndirectUserVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0 // indirect
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	if _, ok := findReplaceVersion(t, gomod, "example.com/dep"); ok {
		t.Fatalf("did not expect replace for indirect user dependency")
	}
}

func TestPinConflictingHookDependenciesIgnoresPreV1Version(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v0.3.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v0.4.0
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	if _, ok := findReplaceVersion(t, gomod, "example.com/dep"); ok {
		t.Fatalf("did not expect replace for pre-v1 dependency")
	}
}

func TestPinConflictingHookDependenciesPreservesExistingReplace(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0

replace example.com/dep => ./localdep
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findReplaceVersion(t, gomod, "example.com/dep")
	if !ok {
		t.Fatalf("expected existing replace for example.com/dep")
	}
	if version != "" {
		t.Fatalf("replace version = %q, want existing local replace without version", version)
	}
}

func TestPinConflictingHookDependenciesIgnoresNonConflictingDependency(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/userdep v1.0.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/hookdep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	if _, ok := findReplaceVersion(t, gomod, "example.com/hookdep"); ok {
		t.Fatalf("did not expect replace for non-conflicting hook dependency")
	}
}
