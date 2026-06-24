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

const meguminnnnnnnnn_openai_dependency_name = "github.com/meguminnnnnnnnn/go-openai"
const meguminnnnnnnnn_openai_module_name = "meguminnnnnnnnn-openai"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("meguminnnnnnnnn-openai-chat-completion-test", meguminnnnnnnnn_openai_module_name,
			"v0.0.0-20250821095446-07791bea23a0", "", "1.22.0", "", TestMeguminnnnnnnnnChatCompletion),
		NewGeneralTestCase("meguminnnnnnnnn-openai-chat-stream-test", meguminnnnnnnnn_openai_module_name,
			"v0.0.0-20250821095446-07791bea23a0", "", "1.22.0", "", TestMeguminnnnnnnnnChatStream),
		NewGeneralTestCase("meguminnnnnnnnn-openai-embeddings-test", meguminnnnnnnnn_openai_module_name,
			"v0.0.0-20250821095446-07791bea23a0", "", "1.22.0", "", TestMeguminnnnnnnnnEmbeddings),
	)
}

func TestMeguminnnnnnnnnChatCompletion(t *testing.T, env ...string) {
	UseApp("meguminnnnnnnnn-openai/v0.0.0-20250821095446")
	RunGoBuild(t, "go", "build", "test_chat_completion.go")
	RunApp(t, "./test_chat_completion", env...)
}

func TestMeguminnnnnnnnnChatStream(t *testing.T, env ...string) {
	UseApp("meguminnnnnnnnn-openai/v0.0.0-20250821095446")
	RunGoBuild(t, "go", "build", "test_chat_completion_stream.go")
	RunApp(t, "./test_chat_completion_stream", env...)
}

func TestMeguminnnnnnnnnEmbeddings(t *testing.T, env ...string) {
	UseApp("meguminnnnnnnnn-openai/v0.0.0-20250821095446")
	RunGoBuild(t, "go", "build", "test_embeddings.go")
	RunApp(t, "./test_embeddings", env...)
}
