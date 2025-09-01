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
	// Test Generate API instrumentation
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Creating default client: %v", err)
		client = &api.Client{} // Create default client for testing
	}
	
	ctx := context.Background()
	req := &api.GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "Hello, world!",
	}
	
	fmt.Println("Testing Generate API instrumentation...")
	
	// Track responses
	var responses []api.GenerateResponse
	
	// This will trigger our instrumentation
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		responses = append(responses, resp)
		if resp.Done {
			fmt.Printf("Final response: %s\n", resp.Response)
			fmt.Printf("Token counts - Input: %d, Output: %d\n", 
				resp.PromptEvalCount, resp.EvalCount)
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("Generate error (expected if no server): %v\n", err)
	}
	
	fmt.Println("Generate test completed!")
}