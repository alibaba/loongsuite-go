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
	"os"
	"testing"

	_ "github.com/alibaba/loongsuite-go-agent/test/version"
)

const test_plugin_name_key = "TEST_PLUGIN_NAME"

func findPlugin() []*TestCase {
	testPluginName := os.Getenv(test_plugin_name_key)
	cases := make([]*TestCase, 0)
	for _, c := range TestCases {
		if c == nil {
			continue
		}
		if c.IsMuzzleCheck || c.IsLatestDepthCheck ||
			(testPluginName != "" && c.TestName != testPluginName) {
			continue
		}
		cases = append(cases, c)
	}
	return cases
}

func findMuzzle() []*TestCase {
	testPluginName := os.Getenv(test_plugin_name_key)
	cases := make([]*TestCase, 0)
	for _, c := range TestCases {
		if c == nil {
			continue
		}
		if !c.IsMuzzleCheck ||
			(testPluginName != "" && c.TestName != testPluginName) {
			continue
		}
		cases = append(cases, c)
	}
	return cases
}

func findLatest() []*TestCase {
	testPluginName := os.Getenv(test_plugin_name_key)
	cases := make([]*TestCase, 0)
	for _, c := range TestCases {
		if c == nil {
			continue
		}
		if !c.IsLatestDepthCheck ||
			(testPluginName != "" && c.TestName != testPluginName) {
			continue
		}
		cases = append(cases, c)
	}
	return cases
}

// Split the plugin tests into 4 parts to avoid too many tests in one run
// This is useful for CI to avoid timeout issues.
func TestPlugins1(t *testing.T) {
	cases := findPlugin()
	for _, c := range cases[:len(cases)/4] {
		t.Run(c.TestName, func(t *testing.T) {
			c.TestFunc(t)
		})
	}
}

func TestPlugins2(t *testing.T) {
	cases := findPlugin()
	for _, c := range cases[len(cases)/4 : len(cases)/2] {
		t.Run(c.TestName, func(t *testing.T) {
			c.TestFunc(t)
		})
	}
}

func TestPlugins3(t *testing.T) {
	cases := findPlugin()
	for _, c := range cases[len(cases)/2 : (len(cases)/2 + len(cases)/4)] {
		t.Run(c.TestName, func(t *testing.T) {
			c.TestFunc(t)
		})
	}
}

func TestPlugins4(t *testing.T) {
	cases := findPlugin()
	for _, c := range cases[(len(cases)/2 + len(cases)/4):] {
		t.Run(c.TestName, func(t *testing.T) {
			c.TestFunc(t)
		})
	}
}

func TestMuzzle(t *testing.T) {
	cases := findMuzzle()
	for _, c := range cases {
		t.Run(c.TestName, func(t *testing.T) {
			ExecMuzzle(t, c.DependencyName, c.ModuleName, c.MinVersion,
				c.MaxVersion, c.MuzzleClasses)
		})
	}
}

func TestLatest(t *testing.T) {
	cases := findLatest()
	for _, c := range cases {
		t.Run(c.TestName, func(t *testing.T) {
			ExecLatestTest(t, c.DependencyName, c.ModuleName, c.MinVersion,
				c.MaxVersion, c.LatestDepthFunc)
		})
	}
}
