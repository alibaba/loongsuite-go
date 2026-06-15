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
	"log"
	"os"
	"time"

	"eino-demo/cmd/einoagent/agent"
	"eino-demo/cmd/einoagent/task"
	"eino-demo/pkg/env"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func init() {
	env.MustHasEnvs("OPENAI_EMBEDDING_MODEL", "OPENAI_API_KEY", "OPENAI_EMBEDDING_BASE_URL", "OPENAI_CHAT_MODEL", "OPENAI_CHAT_BASE_URL")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := server.Default(server.WithHostPorts(":" + port))

	h.Use(LogMiddleware())

	taskGroup := h.Group("/task")
	if err := task.BindRoutes(taskGroup); err != nil {
		log.Fatal("failed to bind task routes:", err)
	}

	agentGroup := h.Group("/agent")
	if err := agent.BindRoutes(agentGroup); err != nil {
		log.Fatal("failed to bind agent routes:", err)
	}

	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Redirect(302, []byte("/agent"))
	})

	h.Spin()
}

func LogMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())
		c.Next(ctx)
		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		log.Printf("[HTTP] %s %s %d %v\n", method, path, statusCode, latency)
	}
}
