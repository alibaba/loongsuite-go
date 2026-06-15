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
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"golang.org/x/net/http2"
)

func setupHttp2() {
	mux := http.NewServeMux()
	mux.HandleFunc("/a", redirectHandler)
	mux.HandleFunc("/b", helloHandler)
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}
	err = http2.ConfigureServer(server, nil)
	if err != nil {
		panic(err)
	}
	err = server.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupHttp2()
	time.Sleep(1 * time.Second)
	jsonData := []byte(`{"key1": "value1", "key2": "value2"}`)
	req, err := http.NewRequest("POST", "https://127.0.0.1:"+strconv.Itoa(port)+"/a", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	err = http2.ConfigureTransport(tr)
	if err != nil {
		panic(err)
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	println(port)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "POST", "POST", "https://127.0.0.1:"+strconv.Itoa(port)+"/a", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "POST /a", "POST", "https", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/2.0", "https", "/a", "", "/a", 200)
		verifier.VerifyHttpClientAttributes(stubs[0][2], "GET", "GET", "http://127.0.0.1:"+strconv.Itoa(port)+"/b", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 400, 0, int64(port))
		if stubs[0][1].Parent.TraceID().String() != stubs[0][0].SpanContext.TraceID().String() {
			log.Fatal("span 1 should be child of span 0")
		}
		if stubs[0][2].Parent.TraceID().String() != stubs[0][1].SpanContext.TraceID().String() {
			log.Fatal("span 2 should be child of span 1")
		}
	}, 1)
}
