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

package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/module"
)

const (
	GoBuildIgnoreComment = "//go:build ignore"
	GoModFile            = "go.mod"
	GoSumFile            = "go.sum"
	GoWorkSumFile        = "go.work.sum"
	DebugLogFile         = "debug.log"
	TempBuildDir         = ".otel-build"
)

const (
	BuildPattern    = "-p"
	BuildImportPath = "-importpath"
	BuildGoVer      = "-goversion"
	BuildPgoProfile = "-pgoprofile"
	BuildModeVendor = "-mod=vendor"
	BuildModeMod    = "-mod=mod"
	BuildWork       = "-work"
)

func AssertGoBuild(args []string) {
	if len(args) < 2 {
		Assert(false, "empty go build command")
	}
	if !strings.Contains(args[0], "go") {
		Assert(false, fmt.Sprintf("invalid go build command %v", args))
	}
	if args[1] != "build" && args[1] != "install" {
		Assert(false, fmt.Sprintf("invalid go build command %v", args))
	}
}

func isToolCommand(tool string, line string, check []string) bool {
	if IsWindows() {
		check = append(check, tool+".exe")
	} else if IsUnix() {
		check = append(check, tool)
	} else {
		ShouldNotReachHere()
	}

	// Check if the line contains all the required fields
	for _, id := range check {
		if !strings.Contains(line, id) {
			return false
		}
	}
	return true
}

func IsCompileCommand(line string) bool {
	return isToolCommand("compile", line, []string{"-o", "-p", "-buildid"})
}

func IsCgoCommand(line string) bool {
	return isToolCommand("cgo", line, []string{"-importpath"})
}
func GetMatchedRuleFile() string {
	const matchedRuleFile = "matched.json"
	return GetTempBuildDirWith(matchedRuleFile)
}

func GetTempBuildDir() string {
	return filepath.Join(TempBuildDir, GetRunPhase().String())
}

func GetTempBuildDirWith(name string) string {
	return filepath.Join(TempBuildDir, name)
}

func GetLogPath(name string) string {
	return filepath.Join(GetTempBuildDir(), name)
}

func GetInstrumentLogPath(name string) string {
	return filepath.Join(TempBuildDir, PInstrument, name)
}

func GetPreprocessLogPath(name string) string {
	return filepath.Join(TempBuildDir, PPreprocess, name)
}

func GetVarNameOfFunc(fn string) string {
	const varDeclSuffix = "Impl"
	// Use strings.ToUpper for the first character to avoid deprecated strings.Title
	if len(fn) > 0 {
		fn = strings.ToUpper(fn[:1]) + fn[1:]
	}
	return fn + varDeclSuffix
}

var packageRegexp = regexp.MustCompile(`(?m)^package\s+\w+`)

func RenamePackage(source, newPkgName string) string {
	source = packageRegexp.ReplaceAllString(source,
		fmt.Sprintf("package %s\n", newPkgName))
	return source
}

func RemoveGoBuildComment(text string) string {
	text = strings.ReplaceAll(text, GoBuildIgnoreComment, "")
	return text
}

func HasGoBuildComment(text string) bool {
	return strings.Contains(text, GoBuildIgnoreComment)
}

// IsModPath checks if the provided module path is valid.
func IsModPath(path string) bool {
	if strings.Contains(path, "@") {
		pathOnly := strings.Split(path, "@")[0]
		return module.CheckPath(pathOnly) == nil
	}
	return module.CheckPath(path) == nil
}

func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

func IsCgo1GoFile(path string) bool {
	return strings.HasSuffix(path, ".cgo1.go")
}

func IsGoModFile(path string) bool {
	return strings.HasSuffix(path, GoModFile)
}

func IsGoSumFile(path string) bool {
	return strings.HasSuffix(path, "go.sum")
}

func IsGoTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

// SplitCompileCmds splits the command line by space, but keep the quoted part as a
// whole. For example, "a b" c will be split into ["a b", "c"].
func SplitCompileCmds(input string) []string {
	var args []string
	var inQuotes bool
	var arg strings.Builder

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == '"' {
			inQuotes = !inQuotes
			continue
		}

		if c == ' ' && !inQuotes {
			if arg.Len() > 0 {
				args = append(args, arg.String())
				arg.Reset()
			}
			continue
		}

		arg.WriteByte(c)
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	// Fix the escaped backslashes on Windows
	if IsWindows() {
		for i, arg := range args {
			args[i] = strings.ReplaceAll(arg, `\\`, `\`)
		}
	}
	return args
}

func FindFlagValue(cmd []string, flag string) string {
	for i, v := range cmd {
		if v == flag {
			return strings.Trim(cmd[i+1], `"`)
		}
	}
	return ""
}
