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
	"log"
	"runtime"
	"strings"
	"testing"

	"github.com/alibaba/loongsuite-go-agent/test/version"
)

// TestCase represents the configuration and metadata for a test case,
// including version constraints, test logic, and specific test types.
type TestCase struct {
	// TestName is the name of the test case.
	TestName string

	// DependencyName is the name of the dependency being tested.
	DependencyName string

	// ModuleName is the name of the module under test.
	ModuleName string

	// MinVersion is the minimum supported version of the module.
	MinVersion *version.Version

	// MaxVersion is the maximum supported version of the module.
	MaxVersion *version.Version

	// MinGoVersion is the minimum Go version required for the test.
	MinGoVersion *version.Version

	// MaxGoVersion is the maximum Go version supported for the test.
	MaxGoVersion *version.Version

	// TestFunc is the function to run the general test logic.
	TestFunc func(t *testing.T, env ...string)

	// LatestDepthFunc is the function to run the latest depth compatibility test.
	LatestDepthFunc func(t *testing.T, env ...string)

	// MuzzleClasses contains the muzzle classes for the muzzle compatibility check.
	MuzzleClasses []string

	// IsMuzzleCheck indicates if the test case is for muzzle compatibility.
	IsMuzzleCheck bool

	// IsLatestDepthCheck indicates if the test case is for the latest depth check.
	IsLatestDepthCheck bool
}

var TestCases = make([]*TestCase, 0)

// NewGeneralTestCase creates a general test case to validate module compatibility
// with specified version constraints and optional test logic.
func NewGeneralTestCase(testName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion string, testFunc func(t *testing.T, env ...string)) *TestCase {
	minVer, err := version.NewVersion(minVersion)
	if minVersion != "" && err != nil {
		log.Printf("Error parsing min version: %v", err)
	}
	maxVer, err := version.NewVersion(maxVersion)
	if maxVersion != "" && err != nil {
		log.Printf("Error parsing max version: %v", err)
	}
	minGoVer, err := version.NewGoVersion(minGoVersion)
	if minGoVersion != "" && err != nil {
		log.Printf("Error parsing min go version: %v", err)
	}
	maxGoVer, err := version.NewGoVersion(maxGoVersion)
	if maxGoVersion != "" && err != nil {
		log.Printf("Error parsing max go version: %v", err)
	}
	goVersion, _ := version.NewGoVersion(strings.ReplaceAll(runtime.Version(), "go", ""))
	if (minGoVer != nil && goVersion.LessThan(minGoVer)) || (maxGoVer != nil && goVersion.GreaterThan(maxGoVer)) {
		log.Printf("TestName is %s, This test does not support go %s, require go [%s, %s]", testName, goVersion.String(), minGoVersion, maxGoVersion)
		return nil
	}
	return &TestCase{
		TestName:           testName,
		ModuleName:         moduleName,
		MinVersion:         minVer,
		MaxVersion:         maxVer,
		MinGoVersion:       minGoVer,
		MaxGoVersion:       maxGoVer,
		TestFunc:           testFunc,
		IsMuzzleCheck:      false,
		IsLatestDepthCheck: false,
	}
}

// NewMuzzleTestCase creates a test case to validate the muzzle compatibility
// of a module with its dependencies.
func NewMuzzleTestCase(testName, dependencyName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion string, muzzleClasses []string) *TestCase {
	c := NewGeneralTestCase(testName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion, nil)
	if c == nil {
		return nil
	}
	c.IsMuzzleCheck = true
	c.DependencyName = dependencyName
	c.MuzzleClasses = muzzleClasses
	return c
}

// NewLatestDepthTestCase creates a test case to validate the compatibility
// of a module's latest depth with its dependencies.
func NewLatestDepthTestCase(testName, dependencyName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion string, latestTestFunc func(t *testing.T, env ...string)) *TestCase {
	c := NewGeneralTestCase(testName, moduleName, minVersion, maxVersion, minGoVersion, maxGoVersion, nil)
	if c == nil {
		return nil
	}
	c.LatestDepthFunc = latestTestFunc
	c.DependencyName = dependencyName
	c.IsLatestDepthCheck = true
	return c
}
