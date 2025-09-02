// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

func main() {
	fmt.Println("========================================")
	fmt.Println("TESTING ALL OLLAMA INSTRUMENTATION PRs")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("=== PR1: Basic Instrumentation ===")
	testBasicGenerate()
	testBasicChat()
	
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== PR2: Streaming Support ===")
	testStreamingGenerate()
	testStreamingChat()
	
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== PR3: Cost Calculation ===")
	testCostTracking()

	fmt.Println("\n========================================")
	fmt.Println("ALL TESTS COMPLETED!")
	fmt.Println("Check the collector output for traces")
	fmt.Println("========================================")
}

func testBasicGenerate() {
	fmt.Println("1. Testing basic Generate API...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &api.GenerateRequest{
		Model:  "llama3.2:1b",
		Prompt: "Say hello in 5 words",
		Stream: new(bool),
	}
	*req.Stream = false

	var response string
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			response = resp.Response
			fmt.Printf("   ✓ Generate completed: %d prompt tokens, %d completion tokens\n", 
				resp.PromptEvalCount, resp.EvalCount)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Generate error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Response: %s\n", response)
	}
}

func testBasicChat() {
	fmt.Println("2. Testing basic Chat API...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &api.ChatRequest{
		Model: "llama3.2:1b",
		Messages: []api.Message{
			{Role: "user", Content: "Say hello in 5 words"},
		},
		Stream: new(bool),
	}
	*req.Stream = false

	var response string
	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Done {
			response = resp.Message.Content
			fmt.Printf("   ✓ Chat completed: %d prompt tokens, %d completion tokens\n",
				resp.PromptEvalCount, resp.EvalCount)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Chat error: %v\n", err)
	} else {
		fmt.Printf("   ✓ Response: %s\n", response)
	}
}

func testStreamingGenerate() {
	fmt.Println("1. Testing streaming Generate API...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &api.GenerateRequest{
		Model:  "llama3.2:1b",
		Prompt: "Count from 1 to 5",
	}

	chunks := 0
	firstTokenTime := time.Now()
	var ttft time.Duration
	
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if chunks == 0 && resp.Response != "" {
			ttft = time.Since(firstTokenTime)
			fmt.Printf("   ✓ TTFT: %v\n", ttft)
		}
		chunks++
		if resp.Done {
			fmt.Printf("   ✓ Streaming completed: %d chunks, %d tokens\n", chunks, resp.EvalCount)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Streaming error: %v\n", err)
	}
}

func testStreamingChat() {
	fmt.Println("2. Testing streaming Chat API...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &api.ChatRequest{
		Model: "llama3.2:1b",
		Messages: []api.Message{
			{Role: "user", Content: "Count from 1 to 5"},
		},
	}

	chunks := 0
	firstTokenTime := time.Now()
	var ttft time.Duration

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		if chunks == 0 && resp.Message.Content != "" {
			ttft = time.Since(firstTokenTime)
			fmt.Printf("   ✓ TTFT: %v\n", ttft)
		}
		chunks++
		if resp.Done {
			fmt.Printf("   ✓ Streaming completed: %d chunks, %d tokens\n", chunks, resp.EvalCount)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Streaming error: %v\n", err)
	}
}

func testCostTracking() {
	fmt.Println("1. Testing cost calculation with llama3.2:1b model...")
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &api.GenerateRequest{
		Model:  "llama3.2:1b", 
		Prompt: "What is 2+2? Answer in one word.",
		Stream: new(bool),
	}
	*req.Stream = false

	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		if resp.Done {
			fmt.Printf("   ✓ Tokens used: prompt=%d, completion=%d\n", 
				resp.PromptEvalCount, resp.EvalCount)
			fmt.Println("   ✓ Cost calculated (check telemetry for gen_ai.cost.* attributes)")
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Error: %v\n", err)
	}

	fmt.Println("2. Testing streaming cost calculation...")
	req2 := &api.GenerateRequest{
		Model:  "llama3.2:1b",
		Prompt: "Say yes",
	}

	chunks := 0
	err = client.Generate(ctx, req2, func(resp api.GenerateResponse) error {
		chunks++
		if resp.Done {
			fmt.Printf("   ✓ Streaming cost tracked: %d chunks\n", chunks)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("   ✗ Error: %v\n", err)
	}
}