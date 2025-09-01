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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/instrument"
	"github.com/alibaba/loongsuite-go-agent/tool/preprocess"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

const (
	SubcommandSet     = "set"
	SubcommandGo      = "go"
	SubcommandVersion = "version"
	SubcommandRemix   = "remix"
)

var usage = `Usage: {} <command> [args]
Example:
	{} go build
	{} go install
	{} go build main.go
	{} version
	{} set -verbose -rule=custom.json

Command:
	version    print the version
	set        set the configuration
	go         build the Go application
`

func printUsage() {
	name, _ := util.GetToolName()
	usage = strings.ReplaceAll(usage, "{}", name)
	fmt.Print(usage)
}

func initTempDir() error {
	// All temp directories are prepared before, instrument phase should not
	// create any new directories.
	if util.GetRunPhase() == util.PInstrument {
		return nil
	}

	// Make temp build directory if not exists
	if util.PathNotExists(util.TempBuildDir) {
		err := os.MkdirAll(util.TempBuildDir, 0777)
		if err != nil {
			return ex.Error(err)
		}
	}
	for _, subdir := range []string{util.PPreprocess, util.PInstrument} {
		_ = os.RemoveAll(util.GetTempBuildDirWith(subdir))
		err := os.MkdirAll(util.GetTempBuildDirWith(subdir), 0777)
		if err != nil {
			return ex.Error(err)
		}
	}

	return nil
}

func initEnv() error {
	util.Assert(len(os.Args) >= 2, "no command specified")

	// Determine the run phase
	switch {
	case strings.HasSuffix(os.Args[1], SubcommandGo):
		// otel go build?
		util.SetRunPhase(util.PPreprocess)
	case os.Args[1] == SubcommandRemix:
		// otel remix?
		util.SetRunPhase(util.PInstrument)
	default:
		// do nothing
	}

	// Create temp build directory
	err := initTempDir()
	if err != nil {
		return err
	}

	// Prepare shared configuration
	if util.InPreprocess() || util.InInstrument() {
		err = config.InitConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	err := initEnv()
	if err != nil {
		ex.Fatal(err)
	}

	subcmd := os.Args[1]
	switch subcmd {
	case SubcommandVersion:
		config.PrintVersion()
	case SubcommandSet:
		err = config.Configure()
	case SubcommandGo:
		err = preprocess.Preprocess()
	case SubcommandRemix:
		err = instrument.Instrument()
	default:
		printUsage()
	}
	if err != nil {
		ex.Fatal(err)
	}
}
