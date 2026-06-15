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

package rules

// -----------------------------------------------------------------------------
// Instrumentation Rule
//
// Instrumentation rules are used to define the behavior of the instrumentation
// for a specific function call. The rules are defined in the init() function
// of rule.go in each package directory. The rules are then used by the instrument
// package to generate the instrumentation code. Multiple rules can be defined
// for a single function call, and the rules are executed in the order of their
// priority. The rules are executed
// in the order of their priority, from high to low.
// There are several types of rules for different purposes:
// - InstFuncRule: Instrumentation rule for a specific function call
// - InstStructRule: Instrumentation rule for a specific struct type
// - InstFileRule: Instrumentation rule for a specific file

type InstRule interface {
	GetVersion() string    // GetVersion returns the version of the rule
	GetGoVersion() string  // GetGoVersion returns the go version of the rule
	GetImportPath() string // GetImportPath returns import path of the rule
	GetPath() string       // GetPath returns the local path of the rule
	SetPath(path string)   // SetPath sets the local path of the rule
	String() string        // String returns string representation of rule
}

type InstBaseRule struct {
	// Local path of the rule, it designates where we can found the hook code
	Path string `json:"Path,omitempty"`
	// Version of the rule, e.g. "[1.9.1,1.9.2)" or "", it designates the
	// version range of rule, all other version will not be instrumented
	Version string `json:"Version,omitempty"`
	// Go version of the rule, e.g. "[1.22.0,)" or "", it designates the go
	// version range of rule, all other go version will not be instrumented
	GoVersion string `json:"GoVersion,omitempty"`
	// Import path of the rule, e.g. "github.com/gin-gonic/gin", it designates
	// the import path of rule, all other import path will not be instrumented
	ImportPath string `json:"ImportPath,omitempty"`
}

func (rule *InstBaseRule) GetVersion() string    { return rule.Version }
func (rule *InstBaseRule) GetGoVersion() string  { return rule.GoVersion }
func (rule *InstBaseRule) GetImportPath() string { return rule.ImportPath }
func (rule *InstBaseRule) GetPath() string       { return rule.Path }
func (rule *InstBaseRule) SetPath(path string)   { rule.Path = path }
