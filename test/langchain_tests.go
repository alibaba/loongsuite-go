package test

import (
	"testing"
)

const langchain_dependency_name = "github.com/tmc/langchaingo"
const langchain_module_name = "langchain"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("langchain-0.1.13-embed-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainEmbed),
		NewGeneralTestCase("langchain-0.1.13-chains-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainChains),
		NewGeneralTestCase("langchain-0.1.13-agent-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainAgent),
		NewGeneralTestCase("langchain-0.1.13-llmgenerate-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainLlmGenerate),
		NewGeneralTestCase("langchain-0.1.13-relevantdoc-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainRelevantDocuments),
		NewGeneralTestCase("langchain-0.1.13-llm-openai-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainLLMOpenAi),
		NewGeneralTestCase("langchain-0.1.13-llm-ollama-test", langchain_module_name, "0.1.13", "0.1.13", "1.22.0", "", TestLangchainLLMOllama),
	)

}

func TestLangchainEmbed(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_embed.go", "fake_embed.go")
	RunApp(t, "test_embed", env...)
}
func TestLangchainChains(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_chains.go")
	RunApp(t, "test_chains", env...)
}
func TestLangchainAgent(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_agent.go")
	RunApp(t, "test_agent", env...)
}
func TestLangchainLlmGenerate(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_llm_generate.go")
	RunApp(t, "test_llm_generate", env...)
}
func TestLangchainRelevantDocuments(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_relevant_doc.go", "fake_vector_db.go")
	RunApp(t, "test_relevant_doc", env...)
}
func TestLangchainLLMOpenAi(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_llm_openai.go")
	RunApp(t, "test_llm_openai", env...)
}
func TestLangchainLLMOllama(t *testing.T, env ...string) {
	UseApp("langchain/v0.1.13")
	RunGoBuild(t, "go", "build", "test_llm_ollama.go")
	RunApp(t, "test_llm_ollama", env...)
}
