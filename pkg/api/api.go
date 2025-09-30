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

package api

// -----------------------------------------------------------------------------
// Call Context
//
// The CallContext struct is used to pass information between the OnEnter and
// OnExit callbacks. The SkipCall field is used to skip the function call if set
// to true. Params and ReturnVals holds address of parameters and return values
// of the original function call. Modification of the Params and ReturnVals will
// affect the original function call thus should be used with caution.

// !!! pkg/api/api.go will auto-sync to tool/internal/instrument/api.tmpl
type CallContext interface {
	// Skip the original function call
	SetSkipCall(bool)
	// Check if the original function call should be skipped
	IsSkipCall() bool
	// Set the data field, can be used to pass information between OnEnter&OnExit
	SetData(interface{})
	// Get the data field, can be used to pass information between OnEnter&OnExit
	GetData() interface{}
	// Get the map data field by key
	GetKeyData(key string) interface{}
	// Set the map data field by key
	SetKeyData(key string, val interface{})
	// Has the map data field by key
	HasKeyData(key string) bool
	// Get the original function parameter at index idx
	GetParam(idx int) interface{}
	// Change the original function parameter at index idx
	SetParam(idx int, val interface{})
	// Get the original function return value at index idx
	GetReturnVal(idx int) interface{}
	// Change the original function return value at index idx
	SetReturnVal(idx int, val interface{})
	// Get the original function name
	GetFuncName() string
	// Get the package name of the original function
	GetPackageName() string
	// Number of original function parameters
	GetParamCount() int
	// Number of original function return values
	GetReturnValCount() int
}
