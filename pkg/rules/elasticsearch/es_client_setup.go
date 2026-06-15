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

package elasticsearch

import (
	"context"
	"net/http"
	"os"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

var esInstrumenter = BuildElasticSearchInstrumenter()

type esInnerEnabler struct {
	enabled bool
}

func (g esInnerEnabler) Enable() bool {
	return g.enabled
}

var esEnabler = esInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ELASTICSEARCH_ENABLED") != "false"}

//go:linkname beforeElasticSearchPerform github.com/elastic/go-elasticsearch/v8.beforeElasticSearchPerform
func beforeElasticSearchPerform(call api.CallContext, client *elasticsearch.BaseClient, request *http.Request) {
	if !esEnabler.Enable() {
		return
	}
	var addresses []string
	for _, u := range client.Transport.(*elastictransport.Client).URLs() {
		addresses = append(addresses, u.String())
	}
	op, params := getEsOpAndParams(request)
	er := &esRequest{
		request: request,
		address: strings.Join(addresses, ","),
		op:      op,
		params:  params,
	}
	newCtx := esInstrumenter.Start(request.Context(), er)
	call.SetKeyData("ctx", newCtx)
	call.SetKeyData("request", er)
}

//go:linkname afterElasticSearchPerform github.com/elastic/go-elasticsearch/v8.afterElasticSearchPerform
func afterElasticSearchPerform(call api.CallContext, response *http.Response, err error) {
	if !esEnabler.Enable() {
		return
	}
	newCtx := call.GetKeyData("ctx").(context.Context)
	er := call.GetKeyData("request").(*esRequest)
	esInstrumenter.End(newCtx, er, response, err)
}

func getEsOpAndParams(req *http.Request) (string, []any) {
	if req == nil || req.URL == nil {
		return "UNKNOWN", nil
	}
	path := req.URL.Path
	paths := strings.Split(path, "/")
	if len(paths) <= 1 {
		return "UNKNOWN", nil
	}
	if len(paths) == 2 {
		return strings.ToLower(req.Method), nil
	}
	params := make([]any, len(paths)-2)
	// path[0] should be the index name
	for i := 2; i < len(paths); i++ {
		params[i-2] = paths[i]
	}
	return paths[2], params
}
