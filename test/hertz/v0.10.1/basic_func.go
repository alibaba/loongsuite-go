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
	"errors"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"time"
)

type myTracer struct {
}

func MyMiddleware(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		// pre-handle
		// ...
		fmt.Println("before request")

		req.AppendBodyString("k1=v1&")

		err = next(ctx, req, resp)
		if err != nil {
			return
		}
		// post-handle
		// ...
		fmt.Println("after request")

		return nil
	}
}

func ExceptionMiddleware(next client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		err = next(ctx, req, resp)
		if err == nil {
			return errors.New("exception")
		} else {
			return err
		}
	}
}

func (m myTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	println("myTracer Start")
	return ctx
}

func (m myTracer) Finish(ctx context.Context, c *app.RequestContext) {
	fmt.Println("myTracer finish")
	fmt.Printf("request path1 is %s\n", string(c.GetRequest().Path()))
	fmt.Printf("request path2 is %s\n", c.FullPath())
}

func setupWithException() {
	h := server.Default()
	h.GET("/exception", func(ctx context.Context, c *app.RequestContext) {
		c.Error(errors.New("exception"))
		c.JSON(consts.StatusInternalServerError, utils.H{"message": "pong"})
	})

	h.Spin()
}

func setupWithRoute() {
	h := server.Default(server.WithTracer(myTracer{}))
	// However, this one will match "/hertz/v1/" and "/hertz/v2/send"
	h.GET("/hertz/:version/*action", func(ctx context.Context, c *app.RequestContext) {
		version := c.Param("version")
		action := c.Param("action")
		message := version + " is " + action

		c.String(consts.StatusOK, message)
	})
	h.Spin()
}

func setupWithTracer() {
	h := server.Default(server.WithTracer(myTracer{}))

	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	h.Spin()
}

func GetDeadline() {
	c, err := client.NewClient()
	if err != nil {
		return
	}
	c.Use(MyMiddleware)
	status, body, _ := c.GetDeadline(context.Background(), nil, "http://127.0.0.1:8888/ping", time.Now().Add(1*time.Second))
	fmt.Printf("status=%v body=%v\n", status, string(body))
}

func GetRoute() {
	c, err := client.NewClient()
	if err != nil {
		return
	}
	c.Use(MyMiddleware)
	status, body, _ := c.GetDeadline(context.Background(), nil, "http://127.0.0.1:8888/hertz/v1", time.Now().Add(1*time.Second))
	fmt.Printf("status=%v body=%v\n", status, string(body))
	status, body, _ = c.GetDeadline(context.Background(), nil, "http://127.0.0.1:8888/hertz/v2/send", time.Now().Add(1*time.Second))
	fmt.Printf("status=%v body=%v\n", status, string(body))
}

func GetException() {
	c, err := client.NewClient()
	if err != nil {
		return
	}
	c.Use(ExceptionMiddleware)
	status, body, _ := c.GetDeadline(context.Background(), nil, "http://127.0.0.1:8888/exception", time.Now().Add(1*time.Second))
	fmt.Printf("status=%v body=%v\n", status, string(body))
}

func Do() {
	c, err := client.NewClient()
	if err != nil {
		return
	}
	req := &protocol.Request{}
	res := &protocol.Response{}
	req.SetMethod(consts.MethodGet)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	req.SetRequestURI("http://127.0.0.1:8888/ping")
	err = c.Do(context.Background(), req, res)
	if err != nil {
		return
	}
	fmt.Printf("%v", string(res.Body()))
}
