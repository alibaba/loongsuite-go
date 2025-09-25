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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	err := Newf("a")
	err = Wrapf(err, "b")
	require.Contains(t, err.Error(), "a")
	require.Contains(t, err.Error(), "b")

	err = fmt.Errorf("c")
	err = Wrapf(err, "d")
	require.Contains(t, err.Error(), "c")
	require.Contains(t, err.Error(), "d")
}
