// Copyright (c) 2025 Alibaba Group Holding Ltd.
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
package all

func f1() {}

func f2(a string) {}

type recv int
type recv2 int
type recx int

func (r *recv) f3() {}

func f4() int { return 7632 }

func (r *recv2) f5() int { return 7632 }

func (r *recx) f6() int { return 7632 }

func init() {
	f1()
	f2("shanxi")
	new(recv).f3()
	f4()
	new(recv2).f5()
	new(recx).f6()
}
