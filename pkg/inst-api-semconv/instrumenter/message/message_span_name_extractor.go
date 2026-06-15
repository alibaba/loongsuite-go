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

package message

const temp_destination_name = "(temporary)"

type MessageSpanNameExtractor[REQUEST any, RESPONSE any] struct {
	Getter        MessageAttrsGetter[REQUEST, RESPONSE]
	OperationName MessageOperation
}

func (m *MessageSpanNameExtractor[REQUEST, RESPONSE]) Extract(request REQUEST) string {
	destinationName := ""
	if m.Getter.IsTemporaryDestination(request) {
		destinationName = temp_destination_name
	} else {
		destinationName = m.Getter.GetDestination(request)
	}
	if destinationName == "" {
		destinationName = "unknown"
	}
	if m.OperationName != "" {
		destinationName = destinationName + " " + string(m.OperationName)
	}
	return destinationName
}
