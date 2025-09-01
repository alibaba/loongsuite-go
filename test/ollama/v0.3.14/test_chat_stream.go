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
	"time"
	
	"github.com/ollama/ollama/api"
)

const progressReportInterval = 10 // Report progress every 10 chunks

func main() {
	// Test Chat API with streaming enabled
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Creating default client: %v", err)
		client = &api.Client{} // Create default client for testing
	}
	
	ctx := context.Background()
	
	// Explicitly enable streaming
	streamFlag := true
	req := &api.ChatRequest{
		Model: "tinyllama",
		Messages: []api.Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that writes concise responses.",
			},
			{
				Role:    "user",
				Content: "Explain what OpenTelemetry is in 2 sentences.",
			},
		},
		Stream: &streamFlag, // Enable streaming
	}
	
	fmt.Println("Testing Chat API with streaming...")
	fmt.Println("Stream mode: enabled")
	
	// Track streaming metrics
	chunkCount := 0
	firstTokenTime := time.Time{}
	startTime := time.Now()
	var totalContent string
	
	// This will trigger our streaming instrumentation
	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		chunkCount++
		
		// Record first token time
		if chunkCount == 1 && resp.Message.Content != "" {
			firstTokenTime = time.Now()
			ttft := firstTokenTime.Sub(startTime).Milliseconds()
			fmt.Printf("First token received! TTFT: %dms\n", ttft)
		}
		
		// Accumulate content
		totalContent += resp.Message.Content
		
		// Print progress every progressReportInterval chunks
		if chunkCount%progressReportInterval == 0 {
			fmt.Printf("Streaming progress: %d chunks received\n", chunkCount)
		}
		
		// Final chunk
		if resp.Done {
			duration := time.Since(startTime)
			fmt.Printf("\n=== Streaming Complete ===\n")
			fmt.Printf("Total chunks: %d\n", chunkCount)
			fmt.Printf("Total duration: %v\n", duration)
			fmt.Printf("Content length: %d characters\n", len(totalContent))
			fmt.Printf("Token counts - Input: %d, Output: %d\n", 
				resp.PromptEvalCount, resp.EvalCount)
			
			if resp.EvalCount > 0 && duration.Seconds() > 0 {
				tokensPerSecond := float64(resp.EvalCount) / duration.Seconds()
				fmt.Printf("Tokens per second: %.2f\n", tokensPerSecond)
			}
			
			fmt.Printf("\nAssistant response:\n%s\n", totalContent)
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("Chat error (expected if no server): %v\n", err)
	}
	
	fmt.Println("\nChat streaming test completed!")
}