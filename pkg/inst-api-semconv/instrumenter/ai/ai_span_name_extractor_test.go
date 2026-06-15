package ai

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testRequest2 struct {
	System    string
	Operation string
}

func TestAINameExtractor(t *testing.T) {
	Extractor := AISpanNameExtractor[testRequest, any]{
		Getter: commonRequest{},
	}
	spanName := Extractor.Extract(testRequest{Operation: "llm", System: "langchain"})
	assert.Equal(t, "llm", spanName)
}
