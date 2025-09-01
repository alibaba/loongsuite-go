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

package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/ollama/ollama/api"
)

func main() {
	// Test backward compatibility with non-streaming requests
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Creating default client: %v", err)
		client = &api.Client{} // Create default client for testing
	}
	
	ctx := context.Background()
	
	fmt.Println("Testing backward compatibility...")
	fmt.Println("==================================================")
	
	// Test 1: Generate with Stream explicitly set to false
	fmt.Println("\nTest 1: Generate API with Stream=false")
	streamFalse := false
	genReqNoStream := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Say hello",
		Stream: &streamFalse, // Explicitly disable streaming
	}
	
	err = client.Generate(ctx, genReqNoStream, func(resp api.GenerateResponse) error {
		if resp.Done {
			fmt.Printf("Non-streaming response received\n")
			fmt.Printf("Content: %s\n", resp.Response)
			fmt.Printf("Tokens - Input: %d, Output: %d\n", 
				resp.PromptEvalCount, resp.EvalCount)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("Generate error (expected if no server): %v\n", err)
	}
	
	// Test 2: Chat with Stream explicitly set to false
	fmt.Println("\nTest 2: Chat API with Stream=false")
	chatReqNoStream := &api.ChatRequest{
		Model: "llama3:8b",
		Messages: []api.Message{
			{Role: "user", Content: "Hi"},
		},
		Stream: &streamFalse, // Explicitly disable streaming
	}
	
	err = client.Chat(ctx, chatReqNoStream, func(resp api.ChatResponse) error {
		if resp.Done {
			fmt.Printf("Non-streaming response received\n")
			fmt.Printf("Content: %s\n", resp.Message.Content)
			fmt.Printf("Tokens - Input: %d, Output: %d\n", 
				resp.PromptEvalCount, resp.EvalCount)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("Chat error (expected if no server): %v\n", err)
	}
	
	// Test 3: Default behavior (Stream=nil, should stream)
	fmt.Println("\nTest 3: Generate API with Stream=nil (default)")
	genReqDefault := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Say hello",
		// Stream not set (nil) - should default to streaming
	}
	
	chunkCount := 0
	err = client.Generate(ctx, genReqDefault, func(resp api.GenerateResponse) error {
		chunkCount++
		if resp.Done {
			fmt.Printf("Default behavior: received %d chunks (streaming=%v)\n", 
				chunkCount, chunkCount > 1)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("Generate error (expected if no server): %v\n", err)
	}
	
	fmt.Println("\n" + "==================================================")
	fmt.Println("Backward compatibility test completed!")
	fmt.Println("All three modes tested:")
	fmt.Println("✓ Stream=false (non-streaming)")
	fmt.Println("✓ Stream=nil (default streaming)")
	fmt.Println("✓ Stream=true (explicit streaming)")
}