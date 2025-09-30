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
	_ "embed"
	"fmt"
	"go/parser"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

const (
	TJumpLabel      = "/* TRAMPOLINE_JUMP_IF */"
	OtelGlobalsFile = "otel.globals.go"
)

func (rp *RuleProcessor) parseAst(filePath string) (*dst.File, error) {
	file := rp.tryRelocated(filePath)
	rp.parser = ast.NewAstParser()
	var err error
	rp.target, err = rp.parser.ParseFile(file, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return rp.target, nil
}

func (rp *RuleProcessor) writeInstrumented(filePath string, root *dst.File) (string, error) {
	rp.parser = nil
	rp.target = nil
	filePath = rp.tryRelocated(filePath)
	name := filepath.Base(filePath)
	newFile, err := ast.WriteFile(root, filepath.Join(rp.workDir, name))
	if err != nil {
		return "", err
	}
	err = rp.replaceCompileArg(newFile, func(arg string) bool {
		return arg == filePath
	})
	if err != nil {
		return "", ex.Wrapf(err, "filePath %s, compileArgs %v, newArg %s",
			filePath, rp.compileArgs, newFile)
	}
	return newFile, nil
}

func makeName(r *rules.InstFuncRule,
	funcDecl *dst.FuncDecl, onEnter bool) string {
	prefix := TrampolineOnExitName
	if onEnter {
		prefix = TrampolineOnEnterName
	}
	return fmt.Sprintf("%s_%s%s",
		prefix, funcDecl.Name.Name, util.Crc32(r.String()))
}

func findJumpPoint(jumpIf *dst.IfStmt) *dst.BlockStmt {
	// Multiple func rules may apply to the same function, we need to find the
	// appropriate jump point to insert trampoline jump.
	if len(jumpIf.Decs.If) == 1 && jumpIf.Decs.If[0] == TJumpLabel {
		// Insert trampoline jump within the else block
		elseBlock := jumpIf.Else.(*dst.BlockStmt)
		if len(elseBlock.List) > 1 {
			// One trampoline jump already exists, recursively find last one
			ifStmt, ok := elseBlock.List[len(elseBlock.List)-1].(*dst.IfStmt)
			util.Assert(ok, "unexpected statement in trampoline-jump-if")
			return findJumpPoint(ifStmt)
		} else {
			// Otherwise, this is the appropriate jump point
			return elseBlock
		}
	}
	return nil
}

func collectReturnValues(funcDecl *dst.FuncDecl) []dst.Expr {
	var retVals []dst.Expr // nil by default
	if retList := funcDecl.Type.Results; retList != nil {
		retVals = make([]dst.Expr, 0)
		// If return values are named, collect their names, otherwise we try to
		// name them manually for further use
		for i, field := range retList.List {
			if field.Names != nil {
				for _, name := range field.Names {
					retVals = append(retVals, dst.NewIdent(name.Name))
				}
			} else {
				retValIdent := dst.NewIdent(fmt.Sprintf("retVal%d", i))
				field.Names = []*dst.Ident{retValIdent}
				retVals = append(retVals, dst.Clone(retValIdent).(*dst.Ident))
			}
		}
	}
	return retVals
}

func collectArguments(funcDecl *dst.FuncDecl) []dst.Expr {
	// Arguments for onEnter trampoline
	args := make([]dst.Expr, 0)
	// Receiver as argument for trampoline func, if any
	if ast.HasReceiver(funcDecl) {
		if recv := funcDecl.Recv.List; recv != nil {
			receiver := recv[0].Names[0].Name
			args = append(args, ast.AddressOf(ast.Ident(receiver)))
		} else {
			util.Unimplemented()
		}
	}
	// Original function arguments as arguments for trampoline func
	for _, field := range funcDecl.Type.Params.List {
		for _, name := range field.Names {
			args = append(args, ast.AddressOf(ast.Ident(name.Name)))
		}
	}
	return args
}

func (rp *RuleProcessor) createTJumpIf(t *rules.InstFuncRule, funcDecl *dst.FuncDecl,
	args []dst.Expr, retVals []dst.Expr,
) *dst.IfStmt {
	varSuffix := util.Crc32(t.String())
	if config.GetConf().Verbose {
		util.Log("varSuffix: %s for %s", varSuffix, t.String())
	}

	// Generate the trampoline-jump-if. N.B. Note that future optimization pass
	// heavily depends on the structure of trampoline-jump-if. Any change in it
	// should be carefully examined.
	onEnterCall := ast.CallTo(makeName(t, funcDecl, true), args)
	onExitCall := ast.CallTo(makeName(t, funcDecl, false), func() []dst.Expr {
		// NB. DST framework disallows duplicated node in the
		// AST tree, we need to replicate the return values
		// as they are already used in return statement above
		clone := make([]dst.Expr, len(retVals)+1)
		clone[0] = ast.Ident(TrampolineCallContextName + varSuffix)
		for i := 1; i < len(clone); i++ {
			clone[i] = ast.AddressOf(retVals[i-1])
		}
		return clone
	}())
	tjumpInit := ast.DefineStmts(
		ast.Exprs(
			ast.Ident(TrampolineCallContextName+varSuffix),
			ast.Ident(TrampolineSkipName+varSuffix),
		),
		ast.Exprs(onEnterCall),
	)
	tjumpCond := ast.Ident(TrampolineSkipName + varSuffix)
	tjumpBody := ast.BlockStmts(
		ast.ExprStmt(onExitCall),
		ast.ReturnStmt(retVals),
	)
	tjumpElse := ast.Block(ast.DeferStmt(onExitCall))
	tjump := ast.IfStmt(tjumpInit, tjumpCond, tjumpBody, tjumpElse)
	// Add this trampoline-jump-if as optimization candidates
	rp.trampolineJumps = append(rp.trampolineJumps, &TJump{
		target: funcDecl,
		ifStmt: tjump,
		rule:   t,
	})
	// Add label for trampoline-jump-if. Note that the label will be cleared
	// during optimization pass, to make it pretty in the generated code
	tjump.Decs.If.Append(TJumpLabel)
	return tjump
}

func (rp *RuleProcessor) insertToFunc(funcDecl *dst.FuncDecl, tjump *dst.IfStmt) {
	found := false
	if len(funcDecl.Body.List) > 0 {
		firstStmt := funcDecl.Body.List[0]
		if ifStmt, ok := firstStmt.(*dst.IfStmt); ok {
			point := findJumpPoint(ifStmt)
			if point != nil {
				point.List = append(point.List, ast.EmptyStmt())
				point.List = append(point.List, tjump)
				found = true
			}
		}
	}
	if !found {
		// Tag the trampoline-jump-if with a special line directive so that
		// debugger can show the correct line number
		tjump.Decs.Before = dst.NewLine
		tjump.Decs.Start.Append("//line <generated>:1")
		pos := rp.parser.FindPosition(funcDecl.Body)
		if len(funcDecl.Body.List) > 0 {
			// It does happens because we may insert raw code snippets at the
			// function entry. These dynamically generated nodes do not have
			// corresponding node positions. We need to keep looking downward
			// until we find a node that contains position information, and then
			// annotate it with a line directive.
			for i := 0; i < len(funcDecl.Body.List); i++ {
				stmt := funcDecl.Body.List[i]
				pos = rp.parser.FindPosition(stmt)
				if !pos.IsValid() {
					continue
				}
				tag := fmt.Sprintf("//line %s", pos.String())
				stmt.Decorations().Before = dst.NewLine
				stmt.Decorations().Start.Append(tag)
			}
		} else {
			pos = rp.parser.FindPosition(funcDecl.Body)
			tag := fmt.Sprintf("//line %s", pos.String())
			empty := ast.EmptyStmt()
			empty.Decs.Before = dst.NewLine
			empty.Decs.Start.Append(tag)
			funcDecl.Body.List = append(funcDecl.Body.List, empty)
		}
		funcDecl.Body.List = append([]dst.Stmt{tjump}, funcDecl.Body.List...)
	}
}

func (rp *RuleProcessor) insertTJump(t *rules.InstFuncRule,
	funcDecl *dst.FuncDecl) error {
	util.Assert(t.OnEnter != "" || t.OnExit != "", "sanity check")

	// Collect return values for the trampoline function
	retVals := collectReturnValues(funcDecl)

	// Collect all arguments for the trampoline function, including the receiver
	// and the original target function arguments
	args := collectArguments(funcDecl)

	// Generate the trampoline-jump-if. The trampoline-jump-if is a conditional
	// jump that jumps to the trampoline function, it looks something like this
	//
	//	if ctx, skip := otel_trampoline_onenter(&arg); skip {
	//	    otel_trampoline_after(ctx, &retval)
	//	    return ...
	//	} else {
	//	    defer otel_trampoline_onexit(ctx, &retval)
	//	    ...
	//	}
	//
	// The trampoline function is just a relay station that properly assembles
	// the context, handles exceptions, etc, and ultimately jumps to the real
	// hook code. By inserting trampoline-jump-if at the target function entry,
	// we can intercept the original function and execute onenter/onexit hooks.
	tjump := rp.createTJumpIf(t, funcDecl, args, retVals)

	// Find if there is already a trampoline-jump-if, insert new tjump if so,
	// otherwise prepend to block body
	rp.insertToFunc(funcDecl, tjump)

	// Trampoline-jump-if ultimately jumps to the trampoline function, which
	// typically has the following form
	//
	//	func otel_trampoline_before(arg) (HookContext, bool) {
	//	    defer func () { /* handle panic */ }()
	//	    // prepare hook context for real hook code
	//	    hookctx := &HookContextImpl_abc{}
	//	    ...
	//	    // Call the real hook code
	//		realHook(ctx, arg)
	//	    return ctx, skip
	//	}
	//
	// It catches any potential panic from the real hook code, and prepare the
	// hook context for the real hook code. Once all preparations are done, it
	// jumps to the real hook code. Note that each trampoline has its own hook
	// context implementation, which is generated dynamically.
	err := rp.createTrampoline(t)
	if err != nil {
		return err
	}
	return nil
}

func (rp *RuleProcessor) insertRaw(r *rules.InstFuncRule, decl *dst.FuncDecl) error {
	util.Assert(r.OnEnter != "" || r.OnExit != "", "sanity check")
	if r.OnEnter != "" {
		// Prepend raw code snippet to function body for onEnter
		p := ast.NewAstParser()
		onEnterSnippet, err := p.ParseSnippet(r.OnEnter)
		if err != nil {
			return err
		}
		decl.Body.List = append(onEnterSnippet, decl.Body.List...)
	}
	if r.OnExit != "" {
		// Use defer func(){ raw_code_snippet }() for onExit
		p := ast.NewAstParser()
		onExitSnippet, err := p.ParseSnippet(
			fmt.Sprintf("defer func(){ %s }()", r.OnExit),
		)
		if err != nil {
			return err
		}
		decl.Body.List = append(onExitSnippet, decl.Body.List...)
	}
	return nil
}

func nameReturnValues(funcDecl *dst.FuncDecl) {
	if funcDecl.Type.Results != nil {
		idx := 0
		for _, field := range funcDecl.Type.Results.List {
			if field.Names == nil {
				name := fmt.Sprintf("retVal%d", idx)
				field.Names = []*dst.Ident{ast.Ident(name)}
				idx++
			}
		}
	}
}

func sortFuncRules(fnRules []*rules.InstFuncRule) []*rules.InstFuncRule {
	sort.SliceStable(fnRules, func(i, j int) bool {
		return fnRules[i].Order < fnRules[j].Order
	})
	return fnRules
}

//go:embed api.tmpl
var templateAPI string

func (rp *RuleProcessor) writeGlobals(pkgName string) error {
	// Prepare trampoline code header
	p := ast.NewAstParser()
	trampoline, err := p.ParseSource("package " + pkgName)
	if err != nil {
		return err
	}
	// Declare common variable declarations
	trampoline.Decls = append(trampoline.Decls, rp.varDecls...)

	// Declare the hook context interface
	api, err := p.ParseSource(templateAPI)
	if err != nil {
		return err
	}
	trampoline.Decls = append(trampoline.Decls, api.Decls...)

	// Write trampoline code to file
	path := filepath.Join(rp.workDir, OtelGlobalsFile)
	trampolineFile, err := ast.WriteFile(trampoline, path)
	if err != nil {
		return err
	}
	rp.addCompileArg(trampolineFile)
	rp.keepForDebug(path)
	return nil
}

func (rp *RuleProcessor) enableLineDirective(filePath string) error {
	text, err := util.ReadFile(filePath)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(".*//line ")
	text = re.ReplaceAllString(text, "//line ")
	// All done, persist to file
	_, err = util.WriteFile(filePath, text)
	if err != nil {
		return err
	}
	return nil
}

func (rp *RuleProcessor) applyFuncRules(bundle *rules.RuleBundle) (err error) {
	// Nothing to do if no func rules
	if len(bundle.File2FuncRules) == 0 {
		return nil
	}
	// Applied all matched func rules, either inserting raw code or inserting
	// our trampoline calls.
	for file, fn2rules := range bundle.File2FuncRules {
		util.Assert(filepath.IsAbs(file), "file path must be absolute")
		astRoot, err := rp.parseAst(file)
		if err != nil {
			return err
		}
		rp.trampolineJumps = make([]*TJump, 0)
		// Since we may generate many functions into the same file, while we don't
		// want to further instrument these functions, we need to make sure that
		// the generated function are excluded from the instrumented file.
		oldDecls := make([]dst.Decl, len(astRoot.Decls))
		copy(oldDecls, astRoot.Decls)
		for fnName, rules := range fn2rules {
			for _, decl := range oldDecls {
				nameAndRecvType := strings.Split(fnName, ",")
				name := nameAndRecvType[0]
				recvType := nameAndRecvType[1]
				if ast.MatchFuncDecl(decl, name, recvType) {
					fnDecl := decl.(*dst.FuncDecl)
					util.Assert(fnDecl.Body != nil, "target func body is empty")
					fnName := fnDecl.Name.Name
					// Save raw function declaration
					rp.rawFunc = fnDecl
					// The func rule can either fully match the target function
					// or use a regexp to match a batch of functions. The
					// generation of tjump differs slightly between these two
					// cases. In the former case, the hook function is required
					// to have the same signature as the target function, while
					// the latter does not have this requirement.
					rp.exact = fnName == name
					// Add explicit names for return values, they can be further
					// referenced if we're willing
					nameReturnValues(fnDecl)

					// Apply all matched rules for this function
					fnRules := sortFuncRules(rules)
					for _, rule := range fnRules {
						if rule.UseRaw {
							err = rp.insertRaw(rule, fnDecl)
						} else {
							err = rp.insertTJump(rule, fnDecl)
						}
						if err != nil {
							return err
						}
						util.Log("Apply func rule %s (%v)", rule, rp.compileArgs)
					}
					// break
				}
			}
		}
		// Optimize generated trampoline-jump-ifs
		err = rp.optimizeTJumps()
		if err != nil {
			return err
		}
		// Write the instrumented AST to new file and replace the original
		// file in the compile command
		newFile, err := rp.writeInstrumented(file, astRoot)
		if err != nil {
			return err
		}
		// Line directive must be placed at the beginning of the line, otherwise
		// it will be ignored by the compiler
		err = rp.enableLineDirective(newFile)
		if err != nil {
			return err
		}
		rp.keepForDebug(newFile)
	}

	err = rp.writeGlobals(bundle.PackageName)
	if err != nil {
		return err
	}
	return nil
}
