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

package auxiliary

func TestSkip() *int {
	val := 3
	return &val
}
func TestSkip2() int {
	return 1024
}
func p1() { /*Test*/ }
func p2() { // Test
}

func p3(arg1 int, arg2 bool, arg3 float64) (int, bool, float64) {
	return arg1, arg2, arg3
}
func TestGetSet(arg1 int, arg2, arg3 bool, arg4 float64, arg5 string,
	arg6 interface{}, arg7, arg8 map[int]bool, arg9 chan int, arg10 []int) (int, bool, bool, float64, string, interface{}, map[int]bool, map[int]bool, chan int, []int) {
	return arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10
}

type GenericRecv[T any] struct{ X T }
type GenericRecv2[T any] struct{ X T }

func (t *GenericRecv[T]) TestGetSetRecv() {}

func (t GenericRecv2[T]) TestGetSetRecv() {}

type Recv struct{ X int }

func (t *Recv) TestGetSetRecv(arg1 int, arg2 float64) (int, float64) {
	return arg1, arg2
}
func OnlyRet() (int, string) {
	return 1024, "gansu"
}
func OnlyArgs(arg1 int, arg2 string) {
	println(arg1, arg2)
}
func NilArg(arg1 *int) {}

type Fn func()

func TargetWithFuncType(fn ...Fn) {}

func Unused(a, _ int) {}
