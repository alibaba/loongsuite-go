package main

import (
	"context"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// Simulate vector database
type fakeVectorDb struct {
}

var _ vectorstores.VectorStore = fakeVectorDb{}

func (fakeVectorDb) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error) {
	return []string{}, nil
}
func (fakeVectorDb) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	return []schema.Document{
		{
			PageContent: "测试",
			Score:       1,
			Metadata:    map[string]any{"key": "value"},
		},
	}, nil
}
