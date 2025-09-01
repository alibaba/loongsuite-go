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
	// Test Generate API with streaming enabled
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Creating default client: %v", err)
		client = &api.Client{} // Create default client for testing
	}
	
	ctx := context.Background()
	
	// Explicitly enable streaming
	streamFlag := true
	req := &api.GenerateRequest{
		Model:  "tinyllama",
		Prompt: "Write a short poem about coding",
		Stream: &streamFlag, // Enable streaming
	}
	
	fmt.Println("Testing Generate API with streaming...")
	fmt.Println("Stream mode: enabled")
	
	// Track streaming metrics
	chunkCount := 0
	firstTokenTime := time.Time{}
	startTime := time.Now()
	var totalContent string
	
	// This will trigger our streaming instrumentation
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		chunkCount++
		
		// Record first token time
		if chunkCount == 1 && resp.Response != "" {
			firstTokenTime = time.Now()
			ttft := firstTokenTime.Sub(startTime).Milliseconds()
			fmt.Printf("First token received! TTFT: %dms\n", ttft)
		}
		
		// Accumulate content
		totalContent += resp.Response
		
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
			
			fmt.Printf("\nGenerated content:\n%s\n", totalContent)
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("Generate error (expected if no server): %v\n", err)
	}
	
	fmt.Println("\nGenerate streaming test completed!")
}