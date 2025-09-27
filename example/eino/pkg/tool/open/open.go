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

package open

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type OpenFileToolImpl struct {
	config *OpenFileToolConfig
}

type OpenFileToolConfig struct {
}

func defaultOpenFileToolConfig(ctx context.Context) (*OpenFileToolConfig, error) {
	config := &OpenFileToolConfig{}
	return config, nil
}

func NewOpenFileTool(ctx context.Context, config *OpenFileToolConfig) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultOpenFileToolConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	t := &OpenFileToolImpl{config: config}
	tn, err = t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

func (of *OpenFileToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("open", "open a file/dir/web url in the system by default application", of.Invoke)
}

func (of *OpenFileToolImpl) Invoke(ctx context.Context, req OpenReq) (res OpenRes, err error) {
	if req.URI == "" {
		res.Message = "uri is required"
		return res, nil
	}

	// if is file or dir, check if exists
	if isFilePath(req.URI) {
		req.URI = strings.TrimPrefix(req.URI, "file:///")
		if _, err := os.Stat(req.URI); err != nil {
			res.Message = fmt.Sprintf("file not exists: %s", req.URI)
			return res, nil
		}
	}

	err = openURI(req.URI)
	if err != nil {
		res.Message = fmt.Sprintf("failed to open %s: %s", req.URI, err.Error())
		return res, nil
	}

	res.Message = fmt.Sprintf("success, open %s", req.URI)
	return res, nil
}

type OpenReq struct {
	URI string `json:"uri" jsonschema:"description=The uri of the file/dir/web url to open"`
}

type OpenRes struct {
	Message string `json:"message" jsonschema:"description=The message of the operation"`
}

func openURI(uri string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", uri)
	case "darwin":
		cmd = exec.Command("open", uri)
	case "linux":
		cmd = exec.Command("xdg-open", uri)
	default:
		return fmt.Errorf("Unsupported Platform")
	}
	return cmd.Run()
}

func isFilePath(path string) bool {
	s, err := url.Parse(path)
	return err == nil && s.Scheme == "file" && s.Path != ""
}
