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
		NewGeneralTestCase("ollama-0.3.14-backward-compat-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaBackwardCompat),
		NewGeneralTestCase("ollama-0.3.14-generate-metrics-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaGenerateMetrics),
		NewGeneralTestCase("ollama-0.3.14-stream-generate-metrics-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaStreamGenerateMetrics),
		NewGeneralTestCase("ollama-0.3.14-chat-metrics-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaChatMetrics),
		NewGeneralTestCase("ollama-0.3.14-stream-chat-metrics-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaStreamChatMetrics),
		NewGeneralTestCase("ollama-0.3.14-embeddings-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaEmbeddings),
		NewGeneralTestCase("ollama-0.3.14-model-management-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaModelManagement),
		NewGeneralTestCase("ollama-0.3.14-comprehensive-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaComprehensive),
		NewGeneralTestCase("ollama-0.3.14-options-extraction-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaOptionsExtraction),
		NewGeneralTestCase("ollama-0.3.14-server-address-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaServerAddress),
		NewGeneralTestCase("ollama-0.3.14-standard-attributes-test", ollama_module_name, "0.3.14", "0.3.14", "1.22", "", TestOllamaStandardAttributes),
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

func TestOllamaBackwardCompat(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_backward_compat.go", "ollama_common.go")
	RunApp(t, "test_backward_compat", env...)
}

func TestOllamaGenerateMetrics(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_generate_metrics.go", "ollama_common.go")
	RunApp(t, "test_ollama_generate_metrics", env...)
}

func TestOllamaStreamGenerateMetrics(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_stream_generate_metrics.go", "ollama_common.go")
	RunApp(t, "test_ollama_stream_generate_metrics", env...)
}

func TestOllamaChatMetrics(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_chat_metrics.go", "ollama_common.go")
	RunApp(t, "test_ollama_chat_metrics", env...)
}

func TestOllamaStreamChatMetrics(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_ollama_stream_chat_metrics.go", "ollama_common.go")
	RunApp(t, "test_ollama_stream_chat_metrics", env...)
}

func TestOllamaEmbeddings(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_embeddings.go", "ollama_common.go")
	RunApp(t, "test_embeddings", env...)
}

func TestOllamaModelManagement(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_model_management.go", "ollama_common.go")
	RunApp(t, "test_model_management", env...)
}

func TestOllamaComprehensive(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_comprehensive.go", "ollama_common.go")
	RunApp(t, "test_comprehensive", env...)
}

func TestOllamaOptionsExtraction(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_options_extraction.go", "ollama_common.go")
	RunApp(t, "test_options_extraction", env...)
}

func TestOllamaServerAddress(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_server_address.go", "ollama_common.go")
	RunApp(t, "test_server_address", env...)
}

func TestOllamaStandardAttributes(t *testing.T, env ...string) {
	UseApp("ollama/v0.3.14")
	RunGoBuild(t, "go", "build", "test_standard_attributes.go", "ollama_common.go")
	RunApp(t, "test_standard_attributes", env...)
}