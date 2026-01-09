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

const openai_community_dependency_name = "github.com/sashabaranov/go-openai"
const openai_official_dependency_name = "github.com/openai/openai-go"
const openai_community_module_name = "openai"
const openai_official_module_name = "openai-official"

func init() {
	// Community SDK tests (go-openai)
	tc1 := NewGeneralTestCase("openai-community-chat-completion-test", openai_community_module_name, "v1.36.1", "", "1.22.0", "", TestOpenAICommunitySDKChatCompletion)
	tc2 := NewGeneralTestCase("openai-community-chat-stream-test", openai_community_module_name, "v1.36.1", "", "1.22.0", "", TestOpenAICommunitySDKChatStream)
	tc3 := NewMuzzleTestCase("openai-community-muzzle-test", openai_community_dependency_name, openai_community_module_name, "v1.36.1", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"})
	
	// Official SDK tests (openai-go)
	// Note: Streaming tests are skipped due to limitations with instrumenting generic types
	tc4 := NewGeneralTestCase("openai-official-chat-completion-test", openai_official_module_name, "v1.5.0", "", "1.22.0", "", TestOpenAIOfficialSDKChatCompletion)
	tc5 := NewMuzzleTestCase("openai-official-muzzle-test", openai_official_dependency_name, openai_official_module_name, "v1.5.0", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"})
	
	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
	if tc3 != nil {
		TestCases = append(TestCases, tc3)
	}
	if tc4 != nil {
		TestCases = append(TestCases, tc4)
	}
	if tc5 != nil {
		TestCases = append(TestCases, tc5)
	}
}

// Community SDK (sashabaranov/go-openai) tests
func TestOpenAICommunitySDKChatCompletion(t *testing.T, env ...string) {
	UseApp("openai/v1.36.1")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestOpenAICommunitySDKChatStream(t *testing.T, env ...string) {
	UseApp("openai/v1.36.1")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}

// Official SDK (openai/openai-go) tests
func TestOpenAIOfficialSDKChatCompletion(t *testing.T, env ...string) {
	UseApp("openai-official/v1.5.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}
