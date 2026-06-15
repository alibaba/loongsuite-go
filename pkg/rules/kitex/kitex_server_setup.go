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
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
)

//go:linkname beforeNewKitexServerInstrument github.com/cloudwego/kitex/server.beforeNewKitexServerInstrument
func beforeNewKitexServerInstrument(call api.CallContext, opts ...server.Option) {
	if !kitexEnabler.Enable() {
		return
	}
	opts = append(opts, server.WithSuite(newServerSuite()))
	if _, ok := call.GetParam(0).([]server.Option); ok {
		call.SetParam(0, opts)
	}
}

func newServerSuite() *serverSuite {
	serverOpts := server.WithTracer(&serverTracer{})
	cOpts := []server.Option{
		serverOpts,
		server.WithMiddleware(ServerMiddleware()),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
	}
	return &serverSuite{cOpts}
}

type serverSuite struct {
	cOpts []server.Option
}

func (c *serverSuite) Options() []server.Option {
	return c.cOpts
}
