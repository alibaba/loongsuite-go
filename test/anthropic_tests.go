// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

const anthropic_dependency_name = "github.com/anthropics/anthropic-sdk-go"
const anthropic_module_name = "anthropic"

func init() {
	tc1 := NewGeneralTestCase("anthropic-chat-completion-test", anthropic_module_name, "v1.25.0", "", "1.23", "", TestAnthropicSDKChatCompletion)
	tc2 := NewGeneralTestCase("anthropic-chat-stream-test", anthropic_module_name, "v1.25.0", "", "1.23", "", TestAnthropicSDKChatStream)
	tc3 := NewMuzzleTestCase("anthropic-muzzle-test", anthropic_dependency_name, anthropic_module_name, "v1.25.0", "", "1.23", "", []string{"go", "build", "test_chat_completion.go"})

	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
	if tc3 != nil {
		TestCases = append(TestCases, tc3)
	}
}

func TestAnthropicSDKChatCompletion(t *testing.T, env ...string) {
	UseApp("anthropic/v1.25.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestAnthropicSDKChatStream(t *testing.T, env ...string) {
	UseApp("anthropic/v1.25.0")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}
