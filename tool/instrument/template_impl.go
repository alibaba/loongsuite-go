//go:build ignore

// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package instrument

//line <generated>:1
// Seeing is not always believing. The following template is a bit tricky, see
// trampoline.go for more details

// Struct Template
type CallContextImpl struct {
	Params      []interface{}
	ReturnVals  []interface{}
	SkipCall    bool
	Data        interface{}
	FuncName    string
	PackageName string
}

func (c *CallContextImpl) SetSkipCall(skip bool)    { c.SkipCall = skip }
func (c *CallContextImpl) IsSkipCall() bool         { return c.SkipCall }
func (c *CallContextImpl) SetData(data interface{}) { c.Data = data }
func (c *CallContextImpl) GetData() interface{}     { return c.Data }
func (c *CallContextImpl) GetKeyData(key string) interface{} {
	if c.Data == nil {
		return nil
	}
	return c.Data.(map[string]interface{})[key]
}
func (c *CallContextImpl) SetKeyData(key string, val interface{}) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data.(map[string]interface{})[key] = val
}

func (c *CallContextImpl) HasKeyData(key string) bool {
	if c.Data == nil {
		return false
	}
	_, ok := c.Data.(map[string]interface{})[key]
	return ok
}

func (c *CallContextImpl) GetParam(idx int) interface{} {
	switch idx {
	}
	return nil
}
func (c *CallContextImpl) SetParam(idx int, val interface{}) {
	if val == nil {
		c.Params[idx] = nil
		return
	}
	switch idx {
	}
}
func (c *CallContextImpl) GetReturnVal(idx int) interface{} {
	switch idx {
	}
	return nil
}
func (c *CallContextImpl) SetReturnVal(idx int, val interface{}) {
	if val == nil {
		c.ReturnVals[idx] = nil
		return
	}
	switch idx {
	}
}

func (c *CallContextImpl) GetFuncName() string    { return c.FuncName }
func (c *CallContextImpl) GetPackageName() string { return c.PackageName }
func (c *CallContextImpl) GetParamCount() int     { return len(c.Params) }
func (c *CallContextImpl) GetReturnValCount() int { return len(c.ReturnVals) }

// Variable Template
var OtelGetStackImpl func() []byte = nil
var OtelPrintStackImpl func([]byte) = nil

// Trampoline Template
func OtelOnEnterTrampoline() (callContext *CallContextImpl, skip bool) {
	defer func() {
		if err := recover(); err != nil {
			println("failed to exec onEnter hook", "OtelOnEnterNamePlaceholder")
			if e, ok := err.(error); ok {
				println(e.Error())
			}
			fetchStack, printStack := OtelGetStackImpl, OtelPrintStackImpl
			if fetchStack != nil && printStack != nil {
				printStack(fetchStack())
			}
		}
	}()
	callContext = &CallContextImpl{}
	callContext.Params = []interface{}{}
	callContext.FuncName = ""
	callContext.PackageName = ""
	return callContext, callContext.SkipCall
}

func OtelOnExitTrampoline(callContext CallContext) {
	defer func() {
		if err := recover(); err != nil {
			println("failed to exec onExit hook", "OtelOnExitNamePlaceholder")
			if e, ok := err.(error); ok {
				println(e.Error())
			}
			fetchStack, printStack := OtelGetStackImpl, OtelPrintStackImpl
			if fetchStack != nil && printStack != nil {
				printStack(fetchStack())
			}
		}
	}()
	callContext.(*CallContextImpl).ReturnVals = []interface{}{}
}
