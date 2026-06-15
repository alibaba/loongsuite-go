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

// InstRawRule represents a rule that allows raw Go source code injection into
// appropriate target function locations. For example, if we want to inject
// raw code at the entry of target function Bar, we can define a rule:
type InstRawRule struct {
	InstBaseRule
	// The name of the target func to be instrumented
	Func string `json:"func,omitempty"`
	// The name of the receiver type
	Recv string `json:"recv,omitempty"`
	// The raw code to be injected
	Raw string `json:"raw,omitempty"`
}

func (rule *InstRawRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}
