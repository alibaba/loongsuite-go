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

// For testing purpose only
type CallContextImpl struct {
	Params     []interface{}
	ReturnVals []interface{}
	SkipCall   bool
	Data       interface{}
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
	return &c.Params[idx]
}

func (c *CallContextImpl) SetParam(idx int, val interface{}) {
	c.Params[idx] = val
}

func (c *CallContextImpl) GetReturnVal(idx int) interface{} {
	return &c.ReturnVals[idx]
}

func (c *CallContextImpl) SetReturnVal(idx int, val interface{}) {
	c.ReturnVals[idx] = val
}

func (c *CallContextImpl) GetFuncName() string {
	return ""
}

func (c *CallContextImpl) GetPackageName() string {
	return ""
}

func (c *CallContextImpl) GetParamCount() int {
	return len(c.Params)
}

func (c *CallContextImpl) GetReturnValCount() int {
	return len(c.ReturnVals)
}

func NewCallContext() CallContext {
	return &CallContextImpl{
		Params:     make([]interface{}, 1024),
		ReturnVals: make([]interface{}, 1024),
		SkipCall:   false,
	}
}
