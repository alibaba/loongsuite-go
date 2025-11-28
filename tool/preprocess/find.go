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

package preprocess

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

func fixCgoSourcePath(prevLine string, line string) string {
	var baseDir string
	if strings.HasPrefix(prevLine, "cd") {
		baseDir = strings.Split(prevLine, " ")[1]
	}
	re := regexp.MustCompile(`\.\/([a-zA-Z0-9_.-]+\.go)`)
	line = re.ReplaceAllStringFunc(line, func(match string) string {
		filename := strings.TrimPrefix(match, "./")
		newPath := filepath.Join(baseDir, filename)
		return newPath
	})
	return line
}

func recordCgoPath(cgoSources map[string][]string, line string) {
	// Split the current line into single arguments
	args := util.SplitCompileCmds(line)
	workDir := util.FindFlagValue(args, "-I")
	util.Assert(workDir != "", "sanity check")
	// Find the source code file path in the current line and associate it with
	// the work directory
	for _, arg := range args {
		if util.IsGoFile(arg) {
			cgoSources[workDir] = append(cgoSources[workDir], arg)
		}
	}
	util.Log("Recorded cgo sources: %v", cgoSources[workDir])
}

// If the package contains cgo source code, all Go source code files are generated
// during the compilation, it looks something like $WORK/abc/source.cgo1.go
// We should fix the source code file path to the real path for further matching
func fixGoSourcePath(cgoSources map[string][]string, line string) string {
	args := util.SplitCompileCmds(line)
	re := regexp.MustCompile(`^(.*[/\\])([^/\\]+)\.cgo1\.go$`)

	for i, arg := range args {
		if !util.IsCgo1GoFile(arg) {
			continue
		}

		matches := re.FindStringSubmatch(arg)
		if len(matches) < 3 {
			continue
		}

		dirPart := matches[1]
		originalBaseName := matches[2]

		targetFiles, ok := cgoSources[dirPart]
		if !ok {
			continue
		}

		for _, targetFile := range targetFiles {
			targetBaseName := strings.TrimSuffix(filepath.Base(targetFile), ".go")

			if originalBaseName == targetBaseName {
				args[i] = targetFile
				break
			}
		}
	}

	return strings.Join(args, " ")
}

func getCompileCommands() ([]string, error) {
	dryRunLog, err := os.Open(util.GetLogPath(DryRunLog))
	if err != nil {
		return nil, ex.Wrap(err)
	}
	defer func(dryRunLog *os.File) {
		err := dryRunLog.Close()
		if err != nil {
			util.Log("Failed to close dry run log file: %v", err)
		}
	}(dryRunLog)

	// Filter compile commands from dry run log
	compileCmds := make([]string, 0)
	scanner := bufio.NewScanner(dryRunLog)
	// 10MB should be enough to accommodate most long line
	buffer := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buffer, cap(buffer))
	prevLine := ""
	cgoSources := map[string][]string{}
	for scanner.Scan() {
		line := scanner.Text()
		// If it's the compile command, all source code files are included in
		// the line, so we can find the source code easily.
		if util.IsCompileCommand(line) {
			line = strings.Trim(line, " ")
			line = fixGoSourcePath(cgoSources, line)
			util.Log("Fixed go source path: %s", line)
			compileCmds = append(compileCmds, line)
		}
		// If it's the cgo command, we need to concatenate the previous line and
		// the current line to get the correct source code file path.
		if util.IsCgoCommand(line) {
			line = fixCgoSourcePath(prevLine, line)
			recordCgoPath(cgoSources, line)
		}
		prevLine = line
	}
	err = scanner.Err()
	if err != nil {
		return nil, ex.Wrapf(err, "cannot parse dry run log")
	}
	return compileCmds, nil
}

// runDryBuild runs a dry build to get all dependencies needed for the project.
func runDryBuild(goBuildCmd []string) ([]string, error) {
	dryRunLog, err := os.Create(util.GetLogPath(DryRunLog))
	if err != nil {
		return nil, ex.Wrap(err)
	}
	// The full build command is: "go build/install -a -x -n  {...}"
	args := []string{}
	args = append(args, goBuildCmd[:2]...)             // go build/install
	args = append(args, []string{"-a", "-x", "-n"}...) // -a -x -n
	args = append(args, goBuildCmd[2:]...)             // {...} remaining
	util.AssertGoBuild(goBuildCmd)
	util.AssertGoBuild(args)

	// Run the dry build
	util.Log("Run dry build %v", args)
	cmd := exec.Command(args[0], args[1:]...)
	// This is a little anti-intuitive as the error message is not printed to
	// the stderr, instead it is printed to the stdout, only the build tool
	// knows the reason why.
	cmd.Stdout = os.Stdout
	cmd.Stderr = dryRunLog
	// @@Note that dir should not be set, as the dry build should be run in the
	// same directory as the original build command
	cmd.Dir = ""
	err = cmd.Run()
	if err != nil {
		return nil, ex.Wrapf(err, "command %v", args)
	}

	// Find source code lines from dry run log
	sourceCodeLines, err := getCompileCommands()
	if err != nil {
		return nil, err
	}
	return sourceCodeLines, nil
}

func (dp *DepProcessor) findDeps() ([]string, error) {
	// Run a dry build to get all dependencies needed for the project
	// Match the dependencies with available rules and prepare them
	// for the actual instrumentation
	// Run dry build to the build blueprint
	compileCmds, err := runDryBuild(dp.goBuildCmd)
	if err != nil {
		// Tell us more about what happened in the dry run
		errLog, _ := util.ReadFile(util.GetLogPath(DryRunLog))
		return nil, ex.Wrapf(err, "dryRunFail: %s", errLog)
	}
	return compileCmds, nil
}
