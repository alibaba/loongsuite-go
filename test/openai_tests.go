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
	
	// Official SDK tests (openai-go) - v1.5.0
	tc4 := NewGeneralTestCase("openai-official-v1-chat-completion-test", openai_official_module_name, "v1.5.0", "", "1.22.0", "", TestOpenAIOfficialSDKV1ChatCompletion)
	tc5 := NewGeneralTestCase("openai-official-v1-chat-stream-test", openai_official_module_name, "v1.5.0", "", "1.22.0", "", TestOpenAIOfficialSDKV1ChatStream)
	tc6 := NewMuzzleTestCase("openai-official-v1-muzzle-test", openai_official_dependency_name, openai_official_module_name, "v1.5.0", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"})
	
	// Official SDK tests (openai-go) - v2.0.0
	tc7 := NewGeneralTestCase("openai-official-v2-chat-completion-test", openai_official_module_name, "v2.0.0", "", "1.22.0", "", TestOpenAIOfficialSDKV2ChatCompletion)
	tc8 := NewGeneralTestCase("openai-official-v2-chat-stream-test", openai_official_module_name, "v2.0.0", "", "1.22.0", "", TestOpenAIOfficialSDKV2ChatStream)
	tc9 := NewMuzzleTestCase("openai-official-v2-muzzle-test", openai_official_dependency_name, openai_official_module_name, "v2.0.0", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"})
	
	// Official SDK tests (openai-go) - v3.0.0
	tc10 := NewGeneralTestCase("openai-official-v3-chat-completion-test", openai_official_module_name, "v3.0.0", "", "1.22.0", "", TestOpenAIOfficialSDKV3ChatCompletion)
	tc11 := NewGeneralTestCase("openai-official-v3-chat-stream-test", openai_official_module_name, "v3.0.0", "", "1.22.0", "", TestOpenAIOfficialSDKV3ChatStream)
	tc12 := NewMuzzleTestCase("openai-official-v3-muzzle-test", openai_official_dependency_name, openai_official_module_name, "v3.0.0", "", "1.22.0", "", []string{"go", "build", "test_chat_completion.go"})
	
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
	if tc6 != nil {
		TestCases = append(TestCases, tc6)
	}
	if tc7 != nil {
		TestCases = append(TestCases, tc7)
	}
	if tc8 != nil {
		TestCases = append(TestCases, tc8)
	}
	if tc9 != nil {
		TestCases = append(TestCases, tc9)
	}
	if tc10 != nil {
		TestCases = append(TestCases, tc10)
	}
	if tc11 != nil {
		TestCases = append(TestCases, tc11)
	}
	if tc12 != nil {
		TestCases = append(TestCases, tc12)
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

// Official SDK (openai/openai-go) tests - v1.5.0
func TestOpenAIOfficialSDKV1ChatCompletion(t *testing.T, env ...string) {
	UseApp("openai-official/v1.5.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestOpenAIOfficialSDKV1ChatStream(t *testing.T, env ...string) {
	UseApp("openai-official/v1.5.0")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}

// Official SDK (openai/openai-go) tests - v2.0.0
func TestOpenAIOfficialSDKV2ChatCompletion(t *testing.T, env ...string) {
	UseApp("openai-official/v2.0.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestOpenAIOfficialSDKV2ChatStream(t *testing.T, env ...string) {
	UseApp("openai-official/v2.0.0")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}

// Official SDK (openai/openai-go) tests - v3.0.0
func TestOpenAIOfficialSDKV3ChatCompletion(t *testing.T, env ...string) {
	UseApp("openai-official/v3.0.0")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestOpenAIOfficialSDKV3ChatStream(t *testing.T, env ...string) {
	UseApp("openai-official/v3.0.0")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}
