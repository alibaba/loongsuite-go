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

import (
	"encoding/json"
)

// InstFuncRule finds specific function call and instrument by adding new code
type InstFuncRule struct {
	InstBaseRule
	// Function name, e.g. "New"
	Function string `json:"Function,omitempty"`
	// Receiver type name, e.g. "*gin.Engine"
	ReceiverType string `json:"ReceiverType,omitempty"`
	// UseRaw indicates whether to insert raw code string
	UseRaw bool `json:"UseRaw,omitempty"`
	// OnEnter callback, called before original function
	OnEnter string `json:"OnEnter,omitempty"`
	// OnExit callback, called after original function
	OnExit string `json:"OnExit,omitempty"`
	// Dependencies is a list of additional dependencies that must be present
	// for this rule to be applied. All dependencies must exist in the project.
	Dependencies []string `json:"Dependencies,omitempty"`
}

// String returns string representation of the rule
func (rule *InstFuncRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}
