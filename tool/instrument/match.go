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

package instrument

import (
	"encoding/json"
	"os"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

const (
	MatchedRulesJsonFile = "matched_rules.json"
)

func (rp *RuleProcessor) load() ([]*rules.InstRuleSet, error) {
	f := util.GetMatchedRuleFile()
	content, err := os.ReadFile(f)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to read file %s", f)
	}
	rset := make([]*rules.InstRuleSet, 0)
	err = json.Unmarshal(content, &rset)
	if err != nil {
		return nil, ex.Wrapf(err, "failed to unmarshal JSON")
	}
	return rset, nil
}

func (rp *RuleProcessor) match(allSet []*rules.InstRuleSet, args []string) *rules.InstRuleSet {
	// One package can only be matched with one rule set, so it's safe to return
	// the first matched rule set.
	importPath := util.FindFlagValue(args, "-p")
	util.Assert(importPath != "", "sanity check")
	for _, rset := range allSet {
		if rset.ImportPath == importPath {
			return rset
		}
	}
	return nil
}
