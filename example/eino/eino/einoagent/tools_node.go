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

package einoagent

import (
	"context"

	"eino-demo/pkg/tool/gitclone"
	"eino-demo/pkg/tool/open"
	"eino-demo/pkg/tool/task"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

func GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	toolTask, err := NewTaskTool(ctx)
	if err != nil {
		return nil, err
	}

	toolOpen, err := NewOpenFileTool(ctx)
	if err != nil {
		return nil, err
	}

	toolGitClone, err := NewGitCloneFile(ctx)
	if err != nil {
		return nil, err
	}

	toolDDGSearch, err := NewDDGSearch(ctx, nil)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		toolTask,
		toolOpen,
		toolGitClone,
		toolDDGSearch,
	}, nil
}

func defaultDDGSearchConfig(ctx context.Context) (*duckduckgo.Config, error) {
	config := &duckduckgo.Config{}
	return config, nil
}

func NewDDGSearch(ctx context.Context, config *duckduckgo.Config) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultDDGSearchConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	tn, err = duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		return nil, err
	}
	return tn, nil
}

func NewOpenFileTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return open.NewOpenFileTool(ctx, nil)
}

func NewGitCloneFile(ctx context.Context) (tn tool.BaseTool, err error) {
	return gitclone.NewGitCloneFile(ctx, nil)
}

func NewTaskTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return task.NewTaskTool(ctx, nil)
}
