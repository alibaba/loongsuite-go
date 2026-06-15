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

// -----------------------------------------------------------------------------
// Extended Error Handling with Stack Traces
//
// These APIs provide an error handling framework that allows errors to carry
// stack trace information.
//
// The core usage pattern is to create an error with a stack trace at its origin
// This allows the error to be propagated up the call stack by simply returning
// it (return err), without needing to be re-wrapped at each level.
//
// While the simplest pattern is to return the error directly, you can also use
// Wrap or Wrapf at any point in the call chain to add more contextual information.
// This is useful when an intermediate function has valuable context that is not
// available at the error's origin.
//
// Create and wrap an error:
//
// 1. To wrap an existing error from a standard or third-party library, use Wrap
//    or Wrapf. This attaches a stack trace to the original error.
//
//    Example:
//    if err := some_lib.DoSomething(); err != nil {
//        return ex.Wrapf(err, "additional context for the error")
//    }
//    if err := some_lib.DoSomething(); err != nil {
//        return ex.Wrap(err)
//    }
//
// 2. To create a new error from scratch, use Newf. This generates a new
//    error with a stack trace at the point of creation.
//
//    Example:
//    if unexpected {
//        return ex.Newf("unexpected error")
//    }
//
// Terminate the program:
//
// Use Fatalf or Fatal to exit the program with an stackful error. It will print
// the error message and stack trace to the standard error output.

const numSkipFrame = 4 // skip the Errorf/Fatalf caller

// stackfulError represents an error with stack trace information
type stackfulError struct {
	message []string
	frame   []string
	wrapped error
}

func (e *stackfulError) Error() string { return strings.Join(e.message, "\n") }
func (e *stackfulError) Unwrap() error { return e.wrapped }

func getFrames() []string {
	const initFrames = 30
	frameList := make([]string, 0)
	pcs := make([]uintptr, initFrames)
	n := runtime.Callers(numSkipFrame, pcs)
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

// wrapOrCreate wraps an error with stack trace information and a formatted message
// If the error is already a stackfulError, it will be decorated with the new message.
// Otherwise, a new stackfulError will be created.
func wrapOrCreate(previousErr error, format string, args ...any) error {
	se := &stackfulError{}
	if errors.As(previousErr, &se) {
		attach := fmt.Sprintf(format, args...)
		if attach != "" {
			se.message = append(se.message, attach)
		}
		return previousErr
	}
	// User defined error message + existing error message
	errMsg := fmt.Sprintf(format, args...)
	if previousErr != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, previousErr.Error())
	}
	e := &stackfulError{
		message: []string{errMsg},
		frame:   getFrames(),
		wrapped: previousErr,
	}
	return e
}

func Wrap(previousErr error) error {
	return wrapOrCreate(previousErr, "")
}

func Wrapf(previousErr error, format string, args ...any) error {
	return wrapOrCreate(previousErr, format, args...)
}

func New(message string) error {
	return wrapOrCreate(nil, "%s", message)
}
func Newf(format string, args ...any) error {
	return wrapOrCreate(nil, format, args...)
}

func Fatalf(format string, args ...any) {
	Fatal(Newf(format, args...))
}

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
