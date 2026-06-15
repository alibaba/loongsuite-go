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
	"bytes"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupBasicHttp() {
	http.HandleFunc("/a", redirectHandler)
	http.HandleFunc("/b", helloHandler)
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel) // default
	go setupBasicHttp()
	time.Sleep(1 * time.Second)
	_, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/a")
	if err != nil {
		panic(err)
	}
	jsonData := []byte(`{"key1": "value1", "key2": "value2"}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:"+strconv.Itoa(port)+"/a", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:"+strconv.Itoa(port)+"/a", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /a", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/1.1", "http", "/a", "", "/a", 200)
		verifier.VerifyHttpClientAttributes(stubs[0][2], "GET", "GET", "http://127.0.0.1:"+strconv.Itoa(port)+"/b", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][3], "GET /b", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/1.1", "http", "/b", "", "/b", 200)
		if stubs[0][1].Parent.TraceID().String() != stubs[0][0].SpanContext.TraceID().String() {
			log.Fatal("span 1 should be child of span 0")
		}
		if stubs[0][2].Parent.TraceID().String() != stubs[0][1].SpanContext.TraceID().String() {
			log.Fatal("span 2 should be child of span 1")
		}
		if stubs[0][3].Parent.TraceID().String() != stubs[0][2].SpanContext.TraceID().String() {
			log.Fatal("span 3 should be child of span 2")
		}

		verifier.VerifyHttpClientAttributes(stubs[1][0], "POST", "POST", "http://127.0.0.1:"+strconv.Itoa(port)+"/a", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[1][1], "POST /a", "POST", "http", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/1.1", "http", "/a", "", "/a", 200)
		verifier.VerifyHttpClientAttributes(stubs[1][2], "GET", "GET", "http://127.0.0.1:"+strconv.Itoa(port)+"/b", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[1][3], "GET /b", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/1.1", "http", "/b", "", "/b", 200)
		if stubs[1][1].Parent.TraceID().String() != stubs[1][0].SpanContext.TraceID().String() {
			log.Fatal("span 1 should be child of span 0")
		}
		if stubs[1][2].Parent.TraceID().String() != stubs[1][1].SpanContext.TraceID().String() {
			log.Fatal("span 2 should be child of span 1")
		}
		if stubs[1][3].Parent.TraceID().String() != stubs[1][2].SpanContext.TraceID().String() {
			log.Fatal("span 3 should be child of span 2")
		}
	}, 2)
}
