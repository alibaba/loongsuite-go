# GenAI Instrumentation Specification

This document describes the semantic conventions for instrumenting GenAI/LLM applications.

## Span Kinds (`gen_ai.span.kind`)

The `gen_ai.span.kind` attribute categorizes the type of operation in a GenAI application:

| Value | Description |
|-------|-------------|
| `workflow` | Top-level span encompassing an entire LLM application or chain execution |
| `task` | Sub-operation within a workflow, used for nested operations |
| `agent` | Autonomous entity that makes decisions and executes actions based on LLM outputs |
| `tool` | External function or API invocation called by an agent or task |
| `generation` | LLM generation/completion for direct text generation |
| `embedding` | Text-to-vector conversion |
| `retriever` | Document retrieval from vector stores or other sources |
| `reranker` | Reordering/reranking documents based on relevance (reserved) |
