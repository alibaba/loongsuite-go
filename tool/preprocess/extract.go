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

package preprocess

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
)

const (
	unzippedPkgDir = "pkg"
)

func extractGZip(data []byte, targetDir string) error {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return ex.Wrap(err)
	}

	gzReader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return ex.Wrap(err)
	}
	defer func() {
		err := gzReader.Close()
		if err != nil {
			ex.Fatal(err)
		}
	}()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ex.Wrap(err)
		}

		// Skip AppleDouble files (._filename) and other hidden files
		if strings.HasPrefix(filepath.Base(header.Name), "._") ||
			strings.HasPrefix(filepath.Base(header.Name), ".") {
			continue
		}

		// Rename pkg_tmp to pkg in the path
		// Normalize path to Unix style for consistent string operations
		cleanName := filepath.ToSlash(filepath.Clean(header.Name))
		if strings.HasPrefix(cleanName, "pkg_tmp/") {
			cleanName = strings.Replace(cleanName, "pkg_tmp/", "pkg/", 1)
		} else if cleanName == "pkg_tmp" {
			cleanName = unzippedPkgDir
		}

		// Sanitize the file path to prevent Zip Slip vulnerability
		if cleanName == "." || cleanName == ".." ||
			strings.HasPrefix(cleanName, "..") {
			continue
		}

		// Ensure the resolved path is within the target directory
		targetPath := filepath.Join(targetDir, cleanName)
		resolvedPath, err := filepath.EvalSymlinks(targetPath)
		if err != nil {
			// If symlink evaluation fails, use the original path
			resolvedPath = targetPath
		}

		// Check if the resolved path is within the target directory
		relPath, err := filepath.Rel(targetDir, resolvedPath)
		if err != nil || strings.HasPrefix(relPath, "..") ||
			filepath.IsAbs(relPath) {
			continue // Skip files that would be extracted outside target dir
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(targetPath, os.FileMode(header.Mode))
			if err != nil {
				return ex.Wrap(err)
			}

		case tar.TypeReg:
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR,
				os.FileMode(header.Mode))
			if err != nil {
				return ex.Wrap(err)
			}

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return ex.Wrap(err)
			}
			err = file.Close()
			if err != nil {
				return ex.Wrap(err)
			}

		default:
			return ex.Newf("unsupported file type: %c in %s",
				header.Typeflag, header.Name)
		}
	}

	return nil
}
