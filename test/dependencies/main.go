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
	"fmt"
	_ "github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(">>> helloHandler called")
	fmt.Fprintf(w, "Hello, World!\n")
}

func main() {
	ts := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer ts.Close()

	resp, _ := http.Get(ts.URL)
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
}
