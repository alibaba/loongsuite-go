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

package util

import (
	"fmt"
	"os"
	"sync"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
)

var logWriter *os.File = os.Stdout
var logMutex sync.Mutex

var Guarantee = Assert // More meaningful name:)

// Be caution it's not thread safe
func SetLogger(w *os.File) {
	logWriter = w
}

func GetLoggerPath() string {
	return logWriter.Name()
}

func Log(format string, args ...interface{}) {
	template := "[" + GetRunPhase().String() + "] " + format + "\n"
	logMutex.Lock()
	_, err := fmt.Fprintf(logWriter, template, args...)
	if err != nil {
		ex.Fatal(err)
	}
	logMutex.Unlock()
}
