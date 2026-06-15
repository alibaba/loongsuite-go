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

package kitex

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	client "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/transmeta"
)

//go:linkname beforeNewKitexClientInstrument github.com/cloudwego/kitex/client.beforeNewKitexClientInstrument
func beforeNewKitexClientInstrument(call api.CallContext, svcInfo interface{}, opts ...client.Option) {
	if !kitexEnabler.Enable() {
		return
	}
	opts = append(opts, client.WithSuite(newClientSuite()))
	if _, ok := call.GetParam(1).([]client.Option); ok {
		call.SetParam(1, opts)
	}
}

func newClientSuite() *clientSuite {
	clientOpts := client.WithTracer(&clientTracer{})
	cOpts := []client.Option{
		clientOpts,
		client.WithMiddleware(ClientMiddleware()),
		client.WithMetaHandler(transmeta.ClientHTTP2Handler),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	}
	return &clientSuite{cOpts}
}

type clientSuite struct {
	cOpts []client.Option
}

func (c *clientSuite) Options() []client.Option {
	return c.cOpts
}
