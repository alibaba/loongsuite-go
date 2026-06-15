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

package preprocess

import "testing"

func TestMatchVersion(t *testing.T) {
	type args struct {
		version     string
		ruleVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "rule version is not set",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version does not start with 'v'",
			args: args{
				version:     "1.50.3",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "rule version does not contain '['",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "(1.45.0,1.57.1)",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "rule version does not contain '('",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[1.45.0,1.57.1]",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "rule version does not contain ','",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[1.45.0)",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "rule version contains 'v'",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[1.45.0,v1.57.1)",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "rule version contains whitespace",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[ 1.45.0 , 1.57.1 )",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "rule version only sets start",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[1.45.0,)",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "rule version only sets end",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[,1.57.1)",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "rule version sets both start and end",
			args: args{
				version:     "v1.50.3",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version is less than rule version start",
			args: args{
				version:     "v1.44.3",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "version is greater than rule version end",
			args: args{
				version:     "v1.58.0",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "version equals rule version start",
			args: args{
				version:     "v1.45.0",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "version equals rule version end",
			args: args{
				version:     "v1.57.1",
				ruleVersion: "[1.45.0,1.57.1)",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchVersion(tt.args.version, tt.args.ruleVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
