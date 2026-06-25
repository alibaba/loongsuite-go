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

package test

import "testing"

const deepseek_dependency_name = "github.com/cohesion-org/deepseek-go"
const deepseek_module_name = "deepseek"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("deepseek-v1.3.0-chat-completion-test", deepseek_module_name,
			"v1.3.0", "", "1.22.0", "", TestDeepseekChatCompletion),
		NewGeneralTestCase("deepseek-v1.3.0-fim-completion-test", deepseek_module_name,
			"v1.3.0", "", "1.22.0", "", TestDeepseekFIMCompletion),
		NewMuzzleTestCase("deepseek-v1.3.0-muzzle-test", deepseek_dependency_name, deepseek_module_name,
			"v1.3.0", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"}),
	)
}

func TestDeepseekChatCompletion(t *testing.T, env ...string) {
	UseApp("deepseek/v1.3.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestDeepseekFIMCompletion(t *testing.T, env ...string) {
	UseApp("deepseek/v1.3.0")
	RunGoBuild(t, "go", "build", "test_fim_completion.go")
	RunApp(t, "./test_fim_completion", env...)
}
