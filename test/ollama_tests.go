package test

import (
	"testing"
)

const ollama_dependency_name = "github.com/ollama/ollama"
const ollama_module_name = "ollama"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("ollama-0.3.14-all-features-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaAllFeatures),
		NewGeneralTestCase("ollama-0.3.14-invoke-chat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaInvokeChat),
		NewGeneralTestCase("ollama-0.3.14-stream-chat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaStreamChat),
		NewGeneralTestCase("ollama-0.3.14-invoke-generate-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaInvokeGenerate),
		NewGeneralTestCase("ollama-0.3.14-stream-generate-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaStreamGenerate),
		NewGeneralTestCase("ollama-0.3.14-tinyllama-generate-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestTinyLlamaGenerate),
		NewGeneralTestCase("ollama-0.3.14-llama3-chat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestLlama3Chat),
		NewGeneralTestCase("ollama-0.3.14-cost-calculation-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaCostCalculation),
		NewGeneralTestCase("ollama-0.3.14-budget-tracking-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaBudgetTracking),
		NewGeneralTestCase("ollama-0.3.14-backward-compat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaBackwardCompat),
	)
}

func TestOllamaAllFeatures(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_all_features.go", "ollama_common.go")
	RunApp(t, "test_all_features", env...)
}

func TestOllamaInvokeChat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_invoke_chat.go", "ollama_common.go")
	RunApp(t, "test_ollama_invoke_chat", env...)
}

func TestOllamaStreamChat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_stream_chat.go", "ollama_common.go")
	RunApp(t, "test_ollama_stream_chat", env...)
}

func TestOllamaInvokeGenerate(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_invoke_generate.go", "ollama_common.go")
	RunApp(t, "test_ollama_invoke_generate", env...)
}

func TestOllamaStreamGenerate(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_stream_generate.go", "ollama_common.go")
	RunApp(t, "test_ollama_stream_generate", env...)
}

func TestTinyLlamaGenerate(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_tinyllama_generate.go", "ollama_common.go")
	RunApp(t, "test_tinyllama_generate", env...)
}

func TestLlama3Chat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_llama3_chat.go", "ollama_common.go")
	RunApp(t, "test_llama3_chat", env...)
}

func TestOllamaCostCalculation(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_cost_calculation.go", "ollama_common.go")
	RunApp(t, "test_cost_calculation", env...)
}

func TestOllamaBudgetTracking(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_budget_tracking.go", "ollama_common.go")
	RunApp(t, "test_budget_tracking", env...)
}

func TestOllamaBackwardCompat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_backward_compat.go", "ollama_common.go")
	RunApp(t, "test_backward_compat", env...)
}