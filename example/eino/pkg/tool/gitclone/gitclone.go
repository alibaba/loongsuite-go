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

package gitclone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GitCloneFileImpl struct {
	config *GitCloneFileConfig
}

type GitCloneFileConfig struct {
	BaseDir string
}

func defaultGitCloneFileConfig(ctx context.Context) (*GitCloneFileConfig, error) {
	config := &GitCloneFileConfig{
		BaseDir: "./data/repos",
	}
	return config, nil
}

func NewGitCloneFile(ctx context.Context, config *GitCloneFileConfig) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultGitCloneFileConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	if config.BaseDir == "" {
		return nil, fmt.Errorf("base dir cannot be empty")
	}
	t := &GitCloneFileImpl{config: config}
	tn, err = t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

func (g *GitCloneFileImpl) ToEinoTool() (tool.BaseTool, error) {
	return utils.InferTool("gitclone", "git clone or pull a repository", g.Invoke)
}

func (g *GitCloneFileImpl) Invoke(ctx context.Context, req *GitCloneRequest) (res *GitCloneResponse, err error) {
	res = &GitCloneResponse{}

	if req.Url == "" {
		res.Error = "URL cannot be empty"
		return res, nil
	}

	valid, cloneURL := isValidGitURL(req.Url)
	if !valid {
		res.Error = fmt.Sprintf("Invalid Git URL format: %s", req.Url)
		return res, nil
	}

	repoDir, repoName := extractRepoDir(cloneURL)
	repoDir = filepath.Join(g.config.BaseDir, repoDir)
	repoPath := filepath.Join(repoDir, repoName)

	if err := os.MkdirAll(g.config.BaseDir, 0755); err != nil {
		res.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return res, nil
	}

	if req.Action == GitCloneActionClone {
		if _, err := os.Stat(repoPath); err == nil {
			res.Error = "Repository already exists"
			return res, nil
		}

		cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			res.Error = fmt.Sprintf("Clone failed: %v, output: %s", err, output)
			return res, nil
		}
	} else if req.Action == GitCloneActionPull {
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			res.Error = fmt.Sprintf("repo does not exist: %s", repoPath)
			return res, nil
		}

		cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull")
		if output, err := cmd.CombinedOutput(); err != nil {
			res.Error = fmt.Sprintf("Pull failed: %v, output: %s", err, output)
			return res, nil
		}

	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		res.Error = fmt.Sprintf("failed to get absolute [%s] path: %v", repoPath, err)
		return res, nil
	}
	res.Message = fmt.Sprintf("success, repo path: %s", absPath)
	return res, nil
}

// 辅助函数：验证 Git URL 格式
func isValidGitURL(url string) (bool, string) {
	cleanURL := strings.TrimSuffix(url, ".git")

	parts := strings.Split(cleanURL, "/")
	if len(parts) < 2 {
		return false, ""
	}

	var standardURL string
	switch {
	// SSH 格式: git@domain:group/repo
	case strings.HasPrefix(url, "git@"):
		if strings.Contains(url, ":") {
			return true, withGit(url) // 已经是标准 SSH 格式
		}
		return false, ""

	// 完整 HTTPS 格式: https://domain/group/repo
	case strings.HasPrefix(url, "http://"), strings.HasPrefix(url, "https://"):
		return true, withGit(url) // 已经是标准 HTTPS 格式

	default:
		standardURL = "https://" + withGit(url)
	}

	return true, standardURL
}

func withGit(url string) string {
	if !strings.HasSuffix(url, ".git") {
		url += ".git"
	}
	return url
}

// 辅助函数：从 URL 提取 group 和 repo
func extractRepoDir(url string) (string, string) {
	parts := strings.Split(url, "/")
	repoDir := parts[len(parts)-2]
	repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")
	return repoDir, repoName
}

type GitCloneAction string

const (
	GitCloneActionClone GitCloneAction = "clone"
	GitCloneActionPull  GitCloneAction = "pull"
)

type GitCloneRequest struct {
	Url    string         `json:"url" jsonschema:"description=The URL of the repository to clone"`
	Action GitCloneAction `json:"action" jsonschema:"description=The action to perform, 'clone' or 'pull'"`
}

type GitCloneResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
