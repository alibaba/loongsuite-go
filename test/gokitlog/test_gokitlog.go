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
	"net/http"
	"os"
	"time"

	kitlog "github.com/go-kit/log"
)

func hello(w http.ResponseWriter, r *http.Request) {
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stdout))
	logger.Log("go-kit logger")
	w.Write([]byte("hello world"))
}

func main() {
	http.HandleFunc("/hello", hello)
	go func() {
		http.ListenAndServe(":8080", nil)
	}()
	time.Sleep(5 * time.Second)
	client := http.Client{}
	client.Get("http://localhost:8080/hello")
}
