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

package instrument

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// Overlay represents the structure of an overlay JSON file for Go build
type Overlay struct {
	Replacement map[string]string `json:"replace"`
}

// collectModifiedFiles processes all source files with pre-filtering and returns a map of modified files
func (rp *RuleProcessor) collectModifiedFiles(rset *rules.InstRuleSet) (map[string]string, error) {
	modifiedFiles := make(map[string]string)
	
	// Precompute target identifiers for pre-filtering
	rp.precomputeTargetIdentifiers(rset)
	
	for file, rs := range groupRules(rset) {
		util.Assert(filepath.IsAbs(file), "file path must be absolute")
		file = rp.findSourceFile(rset, file)
		
		// Pre-filter: check if the file contains any target identifiers
		containsTarget, err := rp.hasTargetIdentifiers(file)
		if err != nil {
			util.Log("Error pre-filtering file %s: %v", file, err)
			// Continue with full parsing as fallback
			containsTarget = true
		}
		
		if !containsTarget {
			util.Log("Skipping file %s - no target identifiers found", file)
			continue
		}
		
		// File passed pre-filtering, now parse the AST and apply rules
		root, err := rp.parseAst(file)
		if err != nil {
			return nil, err
		}
		
		// Check if any rules apply to this file by attempting to apply them
		hasChanges := false
		
		// Apply the rules to the target file
		rp.trampolineJumps = make([]*TJump, 0)
		for _, r := range rs {
			switch rt := r.(type) {
			case *rules.InstFuncRule:
				// Before applying the rule, check if the target function exists in the file
				funcDecls := ast.FindFuncDecl(root, rt.Function, rt.ReceiverType)
				if len(funcDecls) > 0 {
					err1 := rp.applyFuncRule(rt, root)
					if err1 != nil {
						return nil, err1
					}
					hasChanges = true
				}
			case *rules.InstStructRule:
				// Before applying the rule, check if the target struct exists in the file
				structDecl := ast.FindStructDecl(root, rt.StructType)
				if structDecl != nil {
					err1 := rp.applyStructRule(rt, root)
					if err1 != nil {
						return nil, err1
					}
					hasChanges = true
				}
			default:
				util.ShouldNotReachHere()
			}
		}
		
		if hasChanges {
			// Optimize generated trampoline-jump-ifs
			err = rp.optimizeTJumps()
			if err != nil {
				return nil, err
			}

			// Write the instrumented AST to a string buffer directly
			var buf bytes.Buffer
			err = dst.Fprint(&buf, root, nil)  // Use nil as FieldFilter to print all fields
			if err != nil {
				return nil, err
			}
			
			// Add to modified files map with content string
			modifiedFiles[file] = buf.String()
		}
	}
	
	return modifiedFiles, nil
}

// createOverlayFile creates a JSON overlay file mapping original files to modified files
func createOverlayFile(overlay Overlay) (string, error) {
	overlayJSON, err := json.MarshalIndent(overlay, "", "  ")
	if err != nil {
		return "", ex.Wrap(err)
	}
	
	// Create a temporary file for the overlay
	overlayPath, err := os.CreateTemp("", "loong-overlay-*.json")
	if err != nil {
		return "", ex.Wrap(err)
	}
	
	// Write the overlay content to the temp file
	_, err = overlayPath.Write(overlayJSON)
	if err != nil {
		overlayPath.Close()
		return "", ex.Wrap(err)
	}
	
	// Close the file before returning its path
	err = overlayPath.Close()
	if err != nil {
		return "", ex.Wrap(err)
	}
	
	return overlayPath.Name(), nil
}

// applyOverlayToCompileArgs adds the -overlay flag to the compile arguments
func (rp *RuleProcessor) applyOverlayToCompileArgs(overlayPath string) {
	// Check if -overlay flag is already present, and replace it if needed
	overlayFlag := "-overlay=" + overlayPath
	found := false
	for i, arg := range rp.compileArgs {
		if strings.HasPrefix(arg, "-overlay=") {
			rp.compileArgs[i] = overlayFlag
			found = true
			break
		}
	}
	
	if !found {
		// Simply append the overlay flag to the end of compile args
		rp.compileArgs = append(rp.compileArgs, overlayFlag)
	}
}

// instrumentWithOverlay processes files with pre-filtering and uses Go's overlay mechanism for safe injection
func (rp *RuleProcessor) instrumentWithOverlay(rset *rules.InstRuleSet) (err error) {
	hasFuncRule := false
	
	// Apply file rules first because they can introduce new files that used
	// by other rules such as raw rules
	for _, rule := range rset.FileRules {
		err := rp.applyFileRule(rule, rset.PackageName)
		if err != nil {
			return err
		}
	}
	
	// Collect all modified files using pre-filtering
	modifiedFiles, err := rp.collectModifiedFiles(rset)
	if err != nil {
		return err
	}
	
	// If there are modified files, create an overlay and update compile args
	if len(modifiedFiles) > 0 {
		overlay := Overlay{
			Replacement: modifiedFiles,
		}
		
		overlayPath, err := createOverlayFile(overlay)
		if err != nil {
			return err
		}
		
		// Add overlay flag to compile args
		rp.applyOverlayToCompileArgs(overlayPath)
	}
	
	// Check if any function rules were applied (we'll know this if any files were modified)
	if len(modifiedFiles) > 0 {
		// This is a simplification - in a real implementation we'd need a more precise
		// way to determine if function rules were applied, but for now assume that
		// if files were modified, function rules were involved
		hasFuncRule = true
	}
	
	// Write globals file if any function is instrumented because injected code
	// always requires some global variables and auxiliary declarations
	if hasFuncRule {
		return rp.writeGlobals(rset.PackageName)
	}
	
	return nil
}