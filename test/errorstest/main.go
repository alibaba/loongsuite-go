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

package main

import (
	"errors"
	_ "errorstest/all"
	"errorstest/auxiliary"
	_ "errorstest/dep"
	"fmt"
)

func main() {
	err := errors.New("wow")
	target := errors.Unwrap(err)
	fmt.Printf("%v\n", target)
	ptr := auxiliary.TestSkip()
	fmt.Printf("ptr%v\n", ptr)
	val := auxiliary.TestSkip2()
	fmt.Printf("val%v\n", val)
	arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 := auxiliary.TestGetSet(1, true, false, 3.14, "str", nil, map[int]bool{1: true}, map[int]bool{2: true}, make(chan int), []int{1, 2, 3})
	fmt.Printf("val%v %v %v %v %v %v %v %v %v %v\n", arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	recv := &auxiliary.Recv{}
	a, b := recv.TestGetSetRecv(1, 3.14)
	fmt.Printf("recv%v %v %v\n", recv, a, b)
	auxiliary.OnlyArgs(1, "jiangsu")
	c, d := auxiliary.OnlyRet()
	fmt.Printf("onlyret%v %v\n", c, d)
	auxiliary.NilArg(nil)
}
