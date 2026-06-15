package main

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	calc := new(tools.Calculator)
	ag := agents.NewConversationalAgent(fakeAgentLlm{}, []tools.Tool{getAgeTool{}, calc})
	ex := agents.NewExecutor(ag, agents.WithMaxIterations(100))
	_, err := ex.Call(context.Background(), map[string]any{"input": "请问张三的年龄"})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		// All spans are in one trace
		// Span 0: agentExecutor (root)
		// Span 1: chains
		// Span 2: llmGenerateSingle
		// Span 3: agentAction
		// Span 4: chains
		// Span 5: llmGenerateSingle
		verifier.VerifyLLMCommonAttributesWithGenAISpanKind(stubs[0][0], "agentExecutor", "langchain", trace.SpanKindClient, "agent")
		verifier.VerifyLLMCommonAttributesWithGenAISpanKind(stubs[0][3], "agentAction", "langchain", trace.SpanKindClient, "tool")
	}, 1)
}

type getAgeTool struct {
}

var _ tools.Tool = getAgeTool{}

func (c getAgeTool) Description() string {
	return `这是一个根据人名回答年龄的工具，如果需要查询一个人的年龄可以用这个工具。输入：人名，返回：年龄"`
}
func (c getAgeTool) Name() string {
	return "getAge"
}
func (c getAgeTool) Call(ctx context.Context, input string) (string, error) {
	return "20", nil
}

var step int

type fakeAgentLlm struct {
}

var _ llms.Model = fakeAgentLlm{}

func (f fakeAgentLlm) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	defer func() { step++ }()
	res := new(llms.ContentResponse)
	choice := new(llms.ContentChoice)
	if step == 0 {
		choice.Content = "Do I need to use a tool? Yes\nAction: getAge\nAction Input: 张三"
	} else if step == 1 {
		choice.Content = "No\nAI: 20"
	} else {
		choice.Content = "No\nAI: this not in fake llm"
	}
	fmt.Println(choice)
	res.Choices = append(res.Choices, choice)
	return res, nil
}
func (f fakeAgentLlm) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, f, prompt, options...)
}
