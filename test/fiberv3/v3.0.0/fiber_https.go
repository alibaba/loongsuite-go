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
	"crypto/tls"
	"fmt"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

func requestHttpsServer() {
	client := &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	reqURL := "https://localhost:3000/fiber"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)

	err := client.Do(req, resp)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Response body is:\n%s", resp.Body())
}

func setupHttps() {
	// Initialize a new Fiber app
	app := fiber.New()

	// Define a route for the GET method on the root path '/'
	app.Get("/fiber", func(c fiber.Ctx) error {
		// Send a string response to the client
		return c.SendString("Hello, World 👋!")
	})

	// Start the server on port 3000
	app.Listen(":3000", fiber.ListenConfig{
		CertFile:    "server.crt",
		CertKeyFile: "server.key",
	})

}

func main() {
	// starter server
	go setupHttps()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	requestHttpsServer()
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "https://localhost:3000/fiber", "https", "", "tcp", "ipv4", "", "localhost:3000", 200, 0, 3000)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /fiber", "GET", "https", "tcp", "ipv4", "", "localhost:3000", "fasthttp", "https", "/fiber", "", "/fiber", 200)
	}, 1)
}
