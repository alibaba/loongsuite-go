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

// InstStructRule finds specific struct type and instrument by adding new field
type InstStructRule struct {
	InstBaseRule
	// Struct type name, e.g. "Engine"
	StructType string `json:"StructType,omitempty"`
	// New field name, e.g. "Logger"
	FieldName string `json:"FieldName,omitempty"`
	// New field type, e.g. "zap.Logger"
	FieldType string `json:"FieldType,omitempty"`
}

func (rule *InstStructRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}
