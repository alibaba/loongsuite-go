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
package otel_sdk

type BaggageContainer struct {
	baggage interface{}
}

//go:norace
func (bc *BaggageContainer) TakeSnapShot() interface{} {
	return &BaggageContainer{bc.baggage}
}

func GetBaggageFromGLS() *Baggage {
	gls := GetBaggageContainerFromGLS()
	if gls == nil {
		return nil
	}
	p := gls.(*BaggageContainer).baggage
	if p != nil {
		return p.(*Baggage)
	} else {
		return nil
	}
}

func SetBaggageToGLS(baggage *Baggage) {
	SetBaggageContainerToGLS(&BaggageContainer{baggage})
}

func ClearBaggageInGLS() {
	SetBaggageToGLS(nil)
}

func DeleteBaggageMemberInGLS(key string) bool {
	if bInternal := GetBaggageFromGLS(); bInternal != nil {
		b := bInternal.DeleteMember(key)
		SetBaggageToGLS(&b)
		return true
	}
	return false
}
