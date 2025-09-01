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

package ex

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const numSkipFrame = 3 // skip the Errorf/Fatalf caller

// stackfulError represents an error with stack trace information
type stackfulError struct {
	message []string
	frame   []string
	wrapped error
}

func (e *stackfulError) Error() string { return strings.Join(e.message, "\n") }

func getFrames() []string {
	frameList := make([]string, 0)
	pcs := make([]uintptr, 30)
	n := runtime.Callers(numSkipFrame, pcs[:])
	if n == 0 {
		return frameList
	}
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	cnt := 0
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		const prefix = "github.com/alibaba/loongsuite-go-agent/"
		fnName := strings.TrimPrefix(frame.Function, prefix)
		f := fmt.Sprintf("[%d]%s:%d %s", cnt, frame.File, frame.Line, fnName)
		frameList = append(frameList, f)
		cnt++
	}
	return frameList
}

func Error(previousErr error) error {
	se := &stackfulError{}
	if errors.As(previousErr, &se) {
		se.message = append(se.message, previousErr.Error())
		return previousErr
	}
	e := &stackfulError{
		message: []string{previousErr.Error()},
		frame:   getFrames(),
		wrapped: previousErr,
	}
	return e
}

// Errorf wraps an error with stack trace information and a formatted message
// If you want to decorate the existing error, use it.
func Errorf(previousErr error, format string, args ...any) error {
	se := &stackfulError{}
	if errors.As(previousErr, &se) {
		se.message = append(se.message, fmt.Sprintf(format, args...))
		return previousErr
	}
	e := &stackfulError{
		message: []string{fmt.Sprintf(format, args...)},
		frame:   getFrames(),
		wrapped: previousErr,
	}
	return e
}

func Fatalf(format string, args ...any) { Fatal(Errorf(nil, format, args...)) }

func Fatal(err error) {
	if err == nil {
		panic("Fatal error: unknown")
	}
	e := &stackfulError{}
	if errors.As(err, &e) {
		em := ""
		for i, m := range e.message {
			em += fmt.Sprintf("[%d] %s\n", i, m)
		}
		msg := fmt.Sprintf("Error:\n%s\nStack:\n%s", em, strings.Join(e.frame, "\n"))
		_, _ = fmt.Fprint(os.Stderr, msg)
		os.Exit(1)
	}
	panic(err)
}
