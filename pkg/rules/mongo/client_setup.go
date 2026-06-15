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

package mongo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoInstrumenter = BuildMongoOtelInstrumenter()

type mongoInnerEnabler struct {
	enabled bool
}

func (m mongoInnerEnabler) Enable() bool {
	return m.enabled
}

var mongoEnabler = mongoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_MONGO_ENABLED") != "false"}

//go:linkname mongoOnEnter go.mongodb.org/mongo-driver/mongo.mongoOnEnter
func mongoOnEnter(call api.CallContext, opts ...*options.ClientOptions) {
	if !mongoEnabler.Enable() {
		return
	}
	syncMap := sync.Map{}
	for _, opt := range opts {
		hosts := opt.Hosts
		hostLength := len(hosts)
		if hostLength == 0 {
			continue
		}
		configuredMonitor := opt.Monitor
		var host string
		opt.Monitor = &event.CommandMonitor{
			Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
				if configuredMonitor != nil {
					configuredMonitor.Started(ctx, startedEvent)
				}
				host = hosts[0]
				if hostLength > 1 {
					if infoSplit := strings.Index(startedEvent.ConnectionID, "["); infoSplit > 0 && strings.HasSuffix(startedEvent.ConnectionID, "]") {
						host = startedEvent.ConnectionID[0:infoSplit]
					}
				}
				mongoRequest := mongoRequest{
					CommandName: startedEvent.CommandName,
					Host:        host,
				}
				newCtx := mongoInstrumenter.Start(ctx, mongoRequest)
				syncMap.Store(fmt.Sprintf("%d", startedEvent.RequestID), newCtx)
			},
			Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
				if configuredMonitor != nil {
					configuredMonitor.Succeeded(ctx, succeededEvent)
				}
				if newCtx, ok := syncMap.LoadAndDelete(fmt.Sprintf("%d", succeededEvent.RequestID)); ok && newCtx != nil {
					newContext, ok := newCtx.(context.Context)
					if ok {
						mongoInstrumenter.End(newContext, mongoRequest{
							CommandName: succeededEvent.CommandName,
							Host:        host,
						}, nil, nil)
					}
				}
			},
			Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
				if configuredMonitor != nil {
					configuredMonitor.Failed(ctx, failedEvent)
				}
				if newCtx, ok := syncMap.LoadAndDelete(fmt.Sprintf("%d", failedEvent.RequestID)); ok && newCtx != nil {
					newContext, ok := newCtx.(context.Context)
					if ok {
						mongoInstrumenter.End(newContext, mongoRequest{
							CommandName: failedEvent.CommandName,
							Host:        host,
						}, nil, errors.New(failedEvent.Failure))
					}
				}
			},
		}
	}
}
