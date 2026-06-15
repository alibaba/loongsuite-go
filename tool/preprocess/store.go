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
	"encoding/json"
	"os"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

func (dp *DepProcessor) store(matched []*rules.InstRuleSet) error {
	f := util.GetMatchedRuleFile()
	file, err := os.Create(f)
	if err != nil {
		return ex.Wrapf(err, "failed to create file %s", f)
	}
	defer func() {
		_ = file.Close()
	}()

	bs, err := json.Marshal(matched)
	if err != nil {
		return ex.Wrapf(err, "failed to marshal rules to JSON")
	}

	_, err = file.Write(bs)
	if err != nil {
		return ex.Wrapf(err, "failed to write JSON to file %s", f)
	}
	util.Log("Stored matched sets %s", f)
	return nil
}
