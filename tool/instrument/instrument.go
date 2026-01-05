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
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// -----------------------------------------------------------------------------
// Instrument
//
// The instrument package is used to instrument the source code according to the
// predefined rules. It finds the rules that match the project dependencies and
// applies the rules to the dependencies one by one.

type RuleProcessor struct {
	// The working directory during compilation
	workDir string
	// The target file to be instrumented
	target *dst.File
	// The parser for the target file
	parser *ast.AstParser
	// The compiling arguments for the target file
	compileArgs []string
	// The target function to be instrumented
	targetFunc *dst.FuncDecl
	// Whether the rule is exact match with target function, or it's a regexp match
	exact bool
	// The enter hook function, it should be inserted into the target source file
	onEnterHookFunc *dst.FuncDecl
	// The exit hook function, it should be inserted into the target source file
	onExitHookFunc *dst.FuncDecl
	// Variable declarations waiting to be inserted into target source file
	varDecls []dst.Decl
	// Optimization candidates for the trampoline function
	trampolineJumps []*TJump
	// The declaration of the call context, it should be replenished later
	callCtxDecl *dst.GenDecl
	// The methods of the call context
	callCtxMethods []*dst.FuncDecl
	// Map of target identifiers to check for during pre-filtering
	targetIdentifiers map[string]bool
}

func (rp *RuleProcessor) addDecl(decl dst.Decl) {
	rp.target.Decls = append(rp.target.Decls, decl)
}

func (rp *RuleProcessor) removeDeclWhen(pred func(dst.Decl) bool) dst.Decl {
	for i, decl := range rp.target.Decls {
		if pred(decl) {
			rp.target.Decls = append(rp.target.Decls[:i], rp.target.Decls[i+1:]...)
			return decl
		}
	}
	return nil
}

func (rp *RuleProcessor) addCompileArg(newArg string) {
	rp.compileArgs = append(rp.compileArgs, newArg)
}

func haveSameSuffix(s1, s2 string) bool {
	minLength := len(s1)
	if len(s2) < minLength {
		minLength = len(s2)
	}
	for i := 1; i <= minLength; i++ {
		if s1[len(s1)-i] != s2[len(s2)-i] {
			return false
		}
	}
	return true
}

func (rp *RuleProcessor) keepForDebug(name string) {
	escape := func(s string) string {
		dirName := strings.ReplaceAll(s, "/", "_")
		dirName = strings.ReplaceAll(dirName, ".", "_")
		return dirName
	}
	modPath := util.FindFlagValue(rp.compileArgs, "-p")
	dest := filepath.Join("debug", escape(modPath), filepath.Base(name))
	err := util.CopyFile(name, util.GetInstrumentLogPath(dest))
	if err != nil { // error is tolerable here as this is only for debugging
		util.Log("failed to save debug file %s: %v", dest, err)

	}
}

func groupRules(rset *rules.InstRuleSet) map[string][]rules.InstRule {
	file2rules := make(map[string][]rules.InstRule)
	for file, rules := range rset.FuncRules {
		for _, rule := range rules {
			file2rules[file] = append(file2rules[file], rule)
		}
	}
	for file, rules := range rset.StructRules {
		for _, rule := range rules {
			file2rules[file] = append(file2rules[file], rule)
		}
	}
	return file2rules
}

func (rp *RuleProcessor) findSourceFile(rset *rules.InstRuleSet, file string) string {
	if !rset.HasCgo {
		return file
	}
	base := filepath.Base(file)
	file = strings.TrimSuffix(base, ".go")
	file = file + ".cgo1.go"
	for _, arg := range rp.compileArgs {
		if strings.HasSuffix(arg, file) {
			return arg
		}
	}
	return file
}

// precomputeTargetIdentifiers precomputes the target identifiers for all rules in the rule set
func (rp *RuleProcessor) precomputeTargetIdentifiers(rset *rules.InstRuleSet) {
	rp.targetIdentifiers = make(map[string]bool)
	
	// Process FuncRules
	for _, rules := range rset.FuncRules {
		for _, rule := range rules {
			// Add function name as a target identifier
			rp.targetIdentifiers[rule.Function] = true
			// Add receiver type if present
			if rule.ReceiverType != "" {
				rp.targetIdentifiers[rule.ReceiverType] = true
			}
		}
	}
	
	// Process StructRules
	for _, rules := range rset.StructRules {
		for _, rule := range rules {
			// Add struct type as a target identifier
			rp.targetIdentifiers[rule.StructType] = true
		}
	}
}

// hasTargetIdentifiers performs a quick pre-filtering scan to check if the file contains any target identifiers
func (rp *RuleProcessor) hasTargetIdentifiers(filePath string) (bool, error) {
	// First try with go/scanner.Scanner for token-level scanning
	fset := token.NewFileSet()
	
	// Read the file content
	src, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	// Add file to fset and initialize scanner
	file := fset.AddFile(filePath, fset.Base(), len(src))
	var scan scanner.Scanner
	scan.Init(file, src, nil, scanner.ScanComments)
	
	for {
		_, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		
		// Check if the token is an identifier and matches our targets
		if tok == token.IDENT {
			if _, exists := rp.targetIdentifiers[lit]; exists {
				return true, nil
			}
		}
	}
	
	// If scanner didn't find anything, do a simple string search as fallback
	contentStr := string(src)
	for identifier := range rp.targetIdentifiers {
		if strings.Contains(contentStr, identifier) {
			return true, nil
		}
	}
	
	return false, nil
}

// Enhanced version of the instrument function using overlays
func (rp *RuleProcessor) instrument(rset *rules.InstRuleSet) (err error) {
	// Use the overlay-based approach for instrumentation
	return rp.instrumentWithOverlay(rset)
}

func stripCompleteFlag(args []string) []string {
	for i, arg := range args {
		if arg == "-complete" {
			return append(args[:i], args[i+1:]...)
		}
	}
	return args
}

func interceptCompile(args []string) ([]string, error) {
	util.Assert(util.IsCompileCommand(strings.Join(args, " ")), "sanity check")
	target := util.FindFlagValue(args, "-o")
	util.Assert(target != "", "missing -o flag value")
	// Read compilation output directory
	rp := &RuleProcessor{
		workDir:     filepath.Dir(target),
		target:      nil,
		compileArgs: args,
	}

	// Load matched hook rules from setup phase
	bundles, err := rp.load()
	if err != nil {
		return nil, err
	}

	// Check if the current compile command matches the rules.
	matched := rp.match(bundles, args)
	if matched.IsValid() {
		util.Log("Instrument package %v with %v", matched, args)
		err := rp.instrument(matched)
		if err != nil {
			return nil, err
		}

		// Strip -complete flag as we may insert some hook points that are
		// not ready yet, i.e. they don't have function body
		rp.compileArgs = stripCompleteFlag(rp.compileArgs)
		util.Log("Run instrumented command %s", rp.compileArgs)
	}

	return rp.compileArgs, nil
}

func Toolexec() error {
	// Remove the tool itself from the command line arguments
	args := os.Args[2:]
	// Is compile command?
	if util.IsCompileCommand(strings.Join(args, " ")) {
		var err error
		args, err = interceptCompile(args)
		if err != nil {
			return err
		}
	}
	// Just run the command as is
	return util.RunCmd(args...)
}