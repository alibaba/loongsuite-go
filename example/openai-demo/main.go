// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")

	var client *openai.Client
	if baseURL != "" {
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = baseURL
		client = openai.NewClientWithConfig(config)
	} else {
		client = openai.NewClient(apiKey)
	}

	ctx := context.Background()

	fmt.Println("=== Chat Completion ===")
	chatCompletion(ctx, client)

	fmt.Println("\n=== Streaming Chat Completion ===")
	streamChatCompletion(ctx, client)

	fmt.Println("\nDone!")
}

func chatCompletion(ctx context.Context, client *openai.Client) {
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant."},
			{Role: openai.ChatMessageRoleUser, Content: "What is OpenTelemetry in one sentence?"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		log.Printf("chat completion failed: %v", err)
		return
	}

	if len(resp.Choices) > 0 {
		fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	}
	fmt.Printf("Tokens used: prompt=%d, completion=%d, total=%d\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
}

func streamChatCompletion(ctx context.Context, client *openai.Client) {
	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "Count from 1 to 5."},
		},
		Temperature: 0.7,
		Stream:      true,
	})
	if err != nil {
		log.Printf("stream creation failed: %v", err)
		return
	}
	defer stream.Close()

	fmt.Print("Response: ")
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println()
			break
		}
		if err != nil {
			log.Printf("stream recv failed: %v", err)
			return
		}
		if len(resp.Choices) > 0 {
			fmt.Print(resp.Choices[0].Delta.Content)
		}
	}
}
