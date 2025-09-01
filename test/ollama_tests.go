package test

import (
	"testing"
)

const ollama_dependency_name = "github.com/ollama/ollama"
const ollama_module_name = "ollama"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("ollama-0.3.14-chat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaChat),
		NewGeneralTestCase("ollama-0.3.14-generate-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaGenerate),
		NewGeneralTestCase("ollama-0.3.14-chat-stream-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaChatStream),
		NewGeneralTestCase("ollama-0.3.14-generate-stream-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaGenerateStream),
		NewGeneralTestCase("ollama-0.3.14-backward-compat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaBackwardCompat),
		NewGeneralTestCase("ollama-0.3.14-cost-calculation-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaCostCalculation),
		NewGeneralTestCase("ollama-0.3.14-budget-tracking-test", ollama_module_name, "0.3.14", "0.3.14", "1.22.0", "", TestOllamaBudgetTracking),
	)
}

func TestOllamaChat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_chat.go")
	RunApp(t, "test_chat", env...)
}

func TestOllamaGenerate(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_generate.go")
	RunApp(t, "test_generate", env...)
}

func TestOllamaChatStream(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_chat_stream.go")
	RunApp(t, "test_chat_stream", env...)
}

func TestOllamaGenerateStream(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_generate_stream.go")
	RunApp(t, "test_generate_stream", env...)
}

func TestOllamaBackwardCompat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_backward_compat.go")
	RunApp(t, "test_backward_compat", env...)
}

func TestOllamaCostCalculation(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_cost_calculation.go")
	RunApp(t, "test_cost_calculation", env...)
}

func TestOllamaBudgetTracking(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_budget_tracking.go")
	RunApp(t, "test_budget_tracking", env...)
}