// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consumer

import "github.com/streadway/amqp"

type amqpConsumerTextMapCarrier struct {
	headers amqp.Table
}

func (rtmc amqpConsumerTextMapCarrier) Get(key string) string {
	if v, ok := rtmc.headers[key]; ok {
		if vs, ok2 := v.(string); ok2 {
			return vs
		}
	}
	return ""
}

func (rtmc amqpConsumerTextMapCarrier) Set(key string, value string) {
	rtmc.headers[key] = value
}

func (rtmc amqpConsumerTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(rtmc.headers))
	for k := range rtmc.headers {
		keys = append(keys, k)
	}
	return keys
}
