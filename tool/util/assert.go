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

package util

import (
	"reflect"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
)

func Assert(condition bool, message string) {
	if !condition {
		ex.Fatalf("Assertion failed: %s", message)
	}
}

func AssertType[T any](v any) T {
	value, ok := v.(T)
	if !ok {
		actualType := reflect.TypeOf(v).Name()
		var zero T
		expectType := reflect.TypeOf(zero).String()
		ex.Fatalf("Type assertion failed: %s, expected %s",
			actualType, expectType)
	}
	return value
}

func ShouldNotReachHere() {
	ex.Fatalf("Should not reach here!")
}

func Unimplemented(message string) {
	ex.Fatalf("Unimplemented: %s", message)
}
