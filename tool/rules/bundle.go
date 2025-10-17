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

package rules

import (
	"encoding/json"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

const (
	MatchedRulesJsonFile = "matched_rules.json"
)

// InstRuleSet is a collection of rules that matched with one compilation action
type InstRuleSet struct {
	PackageName string
	ImportPath  string
	FileRules   []*InstFileRule
	FuncRules   map[string][]*InstFuncRule
	StructRules map[string][]*InstStructRule
}

func NewInstRuleSet(importPath string) *InstRuleSet {
	return &InstRuleSet{
		PackageName: "",
		ImportPath:  importPath,
		FileRules:   make([]*InstFileRule, 0),
		FuncRules:   make(map[string][]*InstFuncRule),
		StructRules: make(map[string][]*InstStructRule),
	}
}

func (rb *InstRuleSet) String() string {
	bs, _ := json.Marshal(rb)
	return string(bs)
}

func (rb *InstRuleSet) IsValid() bool {
	return rb != nil &&
		(len(rb.FileRules) > 0 ||
			len(rb.FuncRules) > 0 ||
			len(rb.StructRules) > 0)
}

func (rb *InstRuleSet) AddFuncRule(file string, rule *InstFuncRule) {
	if _, exist := rb.FuncRules[file]; !exist {
		rb.FuncRules[file] = make([]*InstFuncRule, 0)
		rb.FuncRules[file] = []*InstFuncRule{rule}
	} else {
		rb.FuncRules[file] = append(rb.FuncRules[file], rule)
	}
}

func (rb *InstRuleSet) AddStructRule(file string, rule *InstStructRule) {
	if _, exist := rb.StructRules[file]; !exist {
		rb.StructRules[file] = make([]*InstStructRule, 0)
		rb.StructRules[file] = []*InstStructRule{rule}
	} else {
		rb.StructRules[file] = append(rb.StructRules[file], rule)
	}
}

func (rb *InstRuleSet) SetPackageName(name string) {
	rb.PackageName = name
}

func (rb *InstRuleSet) AddFileRule(rule *InstFileRule) {
	rb.FileRules = append(rb.FileRules, rule)
}

func StoreRuleBundles(bundles []*InstRuleSet) error {
	util.GuaranteeInPreprocess()
	ruleFile := util.GetPreprocessLogPath(MatchedRulesJsonFile)
	bs, err := json.Marshal(bundles)
	if err != nil {
		return ex.Wrap(err)
	}
	_, err = util.WriteFile(ruleFile, string(bs))
	if err != nil {
		return err
	}
	return nil
}

func LoadRuleBundles() ([]*InstRuleSet, error) {
	util.GuaranteeInInstrument()

	ruleFile := util.GetPreprocessLogPath(MatchedRulesJsonFile)
	data, err := util.ReadFile(ruleFile)
	if err != nil {
		return nil, err
	}
	var bundles []*InstRuleSet
	err = json.Unmarshal([]byte(data), &bundles)
	if err != nil {
		return nil, ex.Wrapf(err, "bad "+ruleFile)
	}
	return bundles, nil
}
