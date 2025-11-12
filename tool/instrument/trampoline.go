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
	"go/token"
	"strconv"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// -----------------------------------------------------------------------------
// Trampoline Jump
//
// We distinguish between three types of functions: RawFunc, TrampolineFunc, and
// HookFunc. RawFunc is the original function that needs to be instrumented.
// TrampolineFunc is the function that is generated to call the onEnter and
// onExit hooks, it serves as a trampoline to the original function. HookFunc is
// the function that is called at entrypoint and exitpoint of the RawFunc. The
// so-called "Trampoline Jump" snippet is inserted at start of raw func, it is
// guaranteed to be generated within one line to avoid confusing debugging, as
// its name suggests, it jumps to the trampoline function from raw function.
const (
	TrampolineSetParamName           = "SetParam"
	TrampolineGetParamName           = "GetParam"
	TrampolineSetReturnValName       = "SetReturnVal"
	TrampolineGetReturnValName       = "GetReturnVal"
	TrampolineValIdentifier          = "val"
	TrampolineCtxIdentifier          = "c"
	TrampolineParamsIdentifier       = "Params"
	TrampolineFuncNameIdentifier     = "FuncName"
	TrampolinePackageNameIdentifier  = "PackageName"
	TrampolineReturnValsIdentifier   = "ReturnVals"
	TrampolineSkipName               = "skip"
	TrampolineCallContextName        = "callContext"
	TrampolineCallContextType        = "CallContext"
	TrampolineCallContextImplType    = "CallContextImpl"
	TrampolineOnEnterName            = "OtelOnEnterTrampoline"
	TrampolineOnExitName             = "OtelOnExitTrampoline"
	TrampolineOnEnterNamePlaceholder = "\"OtelOnEnterNamePlaceholder\""
	TrampolineOnExitNamePlaceholder  = "\"OtelOnExitNamePlaceholder\""
)

// @@ Modification on this trampoline template should be cautious, as it imposes
// many implicit constraints on generated code, known constraints are as follows:
// - It's performance critical, so it should be as simple as possible
// - It should not import any package because there is no guarantee that package
//   is existed in import config during the compilation, one practical approach
//   is to use function variables and setup these variables in preprocess stage
// - It should not panic as this affects user application
// - Function and variable names are coupled with the framework, any modification
//   on them should be synced with the framework

//go:embed impl.tmpl
var trampolineTemplate string

func (rp *RuleProcessor) materializeTemplate() error {
	// Read trampoline template and materialize onEnter and onExit function
	// declarations based on that
	p := ast.NewAstParser()
	astRoot, err := p.ParseSource(trampolineTemplate)
	if err != nil {
		return err
	}

	rp.varDecls = make([]dst.Decl, 0)
	rp.callCtxMethods = make([]*dst.FuncDecl, 0)
	for _, node := range astRoot.Decls {
		// Materialize function declarations
		if decl, ok := node.(*dst.FuncDecl); ok {
			if decl.Name.Name == TrampolineOnEnterName {
				rp.onEnterHookFunc = decl
				rp.addDecl(decl)
			} else if decl.Name.Name == TrampolineOnExitName {
				rp.onExitHookFunc = decl
				rp.addDecl(decl)
			} else if ast.HasReceiver(decl) {
				// We know exactly this is CallContextImpl method
				t := decl.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name
				util.Assert(t == TrampolineCallContextImplType, "sanity check")
				rp.callCtxMethods = append(rp.callCtxMethods, decl)
				rp.addDecl(decl)
			}
		}
		// Materialize variable declarations
		if decl, ok := node.(*dst.GenDecl); ok {
			// No further processing for variable declarations, just append them
			switch decl.Tok {
			case token.VAR:
				rp.varDecls = append(rp.varDecls, decl)
			case token.TYPE:
				rp.callCtxDecl = decl
				rp.addDecl(decl)
			}
		}
	}
	util.Assert(rp.callCtxDecl != nil, "sanity check")
	util.Assert(len(rp.varDecls) > 0, "sanity check")
	util.Assert(rp.onEnterHookFunc != nil, "sanity check")
	util.Assert(rp.onExitHookFunc != nil, "sanity check")
	return nil
}

func getNames(list *dst.FieldList) []string {
	var names []string
	for _, field := range list.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

func makeOnXName(t *rules.InstFuncRule, onEnter bool) string {
	if onEnter {
		return t.OnEnter
	} else {
		return t.OnExit
	}
}

type ParamTrait struct {
	Index          int
	IsVaradic      bool
	IsInterfaceAny bool
}

func isHookDefined(root *dst.File, rule *rules.InstFuncRule) bool {
	util.Assert(rule.OnEnter != "" || rule.OnExit != "", "hook must be set")
	if rule.OnEnter != "" {
		if ast.FindFuncDeclWithoutRecv(root, rule.OnEnter) == nil {
			return false
		}
	}
	if rule.OnExit != "" {
		if ast.FindFuncDeclWithoutRecv(root, rule.OnExit) == nil {
			return false
		}
	}
	return true
}

func findHookFile(rule *rules.InstFuncRule) (string, error) {
	files, err := listRuleFiles(rule)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if !util.IsGoFile(file) {
			continue
		}
		root, err := ast.ParseFileFast(file)
		if err != nil {
			return "", err
		}
		if isHookDefined(root, rule) {
			return file, nil
		}
	}
	return "", ex.Newf("no hook %s/%s found for %s from %v",
		rule.OnEnter, rule.OnExit, rule.Function, files)
}

func listRuleFiles(rule rules.InstRule) ([]string, error) {
	files, err := util.ListFiles(rule.GetPath())
	if err != nil {
		return nil, err
	}
	switch rule.(type) {
	case *rules.InstFuncRule, *rules.InstFileRule:
		return files, nil
	case *rules.InstStructRule:
		util.ShouldNotReachHereT("insane rule type")
	}
	return nil, nil
}

func getHookFunc(t *rules.InstFuncRule, onEnter bool) (*dst.FuncDecl, error) {
	file, err := findHookFile(t)
	if err != nil {
		return nil, err
	}
	astRoot, err := ast.ParseFile(file)
	if err != nil {
		return nil, err
	}
	var target *dst.FuncDecl
	if onEnter {
		target = ast.FindFuncDeclWithoutRecv(astRoot, t.OnEnter)
	} else {
		target = ast.FindFuncDeclWithoutRecv(astRoot, t.OnExit)
	}
	if target == nil {
		return nil, ex.Newf("hook %s or %s not found", t.OnEnter, t.OnExit)
	}
	return target, nil
}

func getHookParamTraits(t *rules.InstFuncRule, onEnter bool) ([]ParamTrait, error) {
	target, err := getHookFunc(t, onEnter)
	if err != nil {
		return nil, err
	}
	var attrs []ParamTrait
	// Find which parameter is type of interface{}
	for i, field := range target.Type.Params.List {
		attr := ParamTrait{Index: i}
		if ast.IsInterfaceType(field.Type) {
			attr.IsInterfaceAny = true
		}
		if ast.IsEllipsis(field.Type) {
			attr.IsVaradic = true
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

func (rp *RuleProcessor) callOnEnterHook(t *rules.InstFuncRule, traits []ParamTrait) error {
	// The actual parameter list of hook function should be the same as the
	// target function
	if rp.exact {
		util.Assert(len(traits) == (len(rp.onEnterHookFunc.Type.Params.List)+1),
			"hook func signature can not match with target function")
	}
	// Hook: 	   func onEnterFoo(callContext* CallContext, p*[]int)
	// Trampoline: func OtelOnEnterTrampoline_foo(p *[]int)
	args := []dst.Expr{dst.NewIdent(TrampolineCallContextName)}
	if rp.exact {
		for idx, field := range rp.onEnterHookFunc.Type.Params.List {
			trait := traits[idx+1 /*CallContext*/]
			for _, name := range field.Names { // syntax of n1,n2 type
				if trait.IsVaradic {
					args = append(args, ast.DereferenceOf(ast.Ident(name.Name+"...")))
				} else {
					args = append(args, ast.DereferenceOf(dst.NewIdent(name.Name)))
				}
			}
		}
	}
	fnName := makeOnXName(t, true)
	call := ast.ExprStmt(ast.CallTo(fnName, args))
	iff := ast.IfNotNilStmt(
		dst.NewIdent(fnName),
		ast.Block(call),
		nil,
	)
	insertAt(rp.onEnterHookFunc, iff, len(rp.onEnterHookFunc.Body.List)-1)
	return nil
}

func (rp *RuleProcessor) callOnExitHook(t *rules.InstFuncRule, traits []ParamTrait) error {
	// The actual parameter list of hook function should be the same as the
	// target function
	if rp.exact {
		util.Assert(len(traits) == len(rp.onExitHookFunc.Type.Params.List),
			"hook func signature can not match with target function")
	}
	// Hook: 	   func onExitFoo(ctx* CallContext, p*[]int)
	// Trampoline: func OtelOnExitTrampoline_foo(ctx* CallContext, p *[]int)
	var args []dst.Expr
	for idx, field := range rp.onExitHookFunc.Type.Params.List {
		if idx == 0 {
			args = append(args, dst.NewIdent(TrampolineCallContextName))
			if !rp.exact {
				// Generic hook function, no need to process parameters
				break
			}
			continue
		}
		trait := traits[idx]
		for _, name := range field.Names { // syntax of n1,n2 type
			if trait.IsVaradic {
				arg := ast.DereferenceOf(ast.Ident(name.Name + "..."))
				args = append(args, arg)
			} else {
				arg := ast.DereferenceOf(dst.NewIdent(name.Name))
				args = append(args, arg)
			}
		}
	}
	fnName := makeOnXName(t, false)
	call := ast.ExprStmt(ast.CallTo(fnName, args))
	iff := ast.IfNotNilStmt(
		dst.NewIdent(fnName),
		ast.Block(call),
		nil,
	)
	insertAtEnd(rp.onExitHookFunc, iff)
	return nil
}

func rectifyAnyType(paramList *dst.FieldList, traits []ParamTrait) error {
	if len(paramList.List) != len(traits) {
		return ex.Newf("hook func signature can not match with target function")
	}
	for i, field := range paramList.List {
		trait := traits[i]
		if trait.IsInterfaceAny {
			// Rectify type to "interface{}"
			field.Type = ast.InterfaceType()
		}
	}
	return nil
}

func (rp *RuleProcessor) addHookFuncVar(t *rules.InstFuncRule,
	traits []ParamTrait, onEnter bool) error {
	paramTypes := &dst.FieldList{List: []*dst.Field{}}
	if rp.exact {
		paramTypes = rp.buildTrampolineType(onEnter)
	}
	addCallContext(paramTypes)
	if rp.exact {
		// Hook functions may uses interface{} as parameter type, as some types of
		// raw function is not exposed
		err := rectifyAnyType(paramTypes, traits)
		if err != nil {
			return err
		}
	}

	// Generate var decl and append it to the target file, note that many target
	// functions may match the same hook function, it's a fatal error to append
	// multiple hook function declarations to the same file, so we need to check
	// if the hook function variable is already declared in the target file
	exist := false
	fnName := makeOnXName(t, onEnter)
	funcDecl := &dst.FuncDecl{
		Name: &dst.Ident{
			Name: fnName,
		},
		Type: &dst.FuncType{
			Func:   false,
			Params: paramTypes,
		},
	}
	for _, decl := range rp.target.Decls {
		if fDecl, ok := decl.(*dst.FuncDecl); ok {
			if fDecl.Name.Name == fnName {
				exist = true
				break
			}
		}
	}
	if !exist {
		rp.addDecl(funcDecl)
	}
	return nil
}

func insertAt(funcDecl *dst.FuncDecl, stmt dst.Stmt, index int) {
	stmts := funcDecl.Body.List
	newStmts := append(stmts[:index],
		append([]dst.Stmt{stmt}, stmts[index:]...)...)
	funcDecl.Body.List = newStmts
}

func insertAtEnd(funcDecl *dst.FuncDecl, stmt dst.Stmt) {
	insertAt(funcDecl, stmt, len(funcDecl.Body.List))
}

func (rp *RuleProcessor) renameTrampolineFunc(t *rules.InstFuncRule) {
	// Randomize trampoline function names
	rp.onEnterHookFunc.Name.Name = makeName(t, rp.targetFunc, true)
	dst.Inspect(rp.onEnterHookFunc, func(node dst.Node) bool {
		if basicLit, ok := node.(*dst.BasicLit); ok {
			// Replace OtelOnEnterTrampolinePlaceHolder to real hook func name
			if basicLit.Value == TrampolineOnEnterNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnEnter)
			}
		}
		return true
	})
	rp.onExitHookFunc.Name.Name = makeName(t, rp.targetFunc, false)
	dst.Inspect(rp.onExitHookFunc, func(node dst.Node) bool {
		if basicLit, ok := node.(*dst.BasicLit); ok {
			if basicLit.Value == TrampolineOnExitNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnExit)
			}
		}
		return true
	})
}

func addCallContext(list *dst.FieldList) {
	callCtx := ast.NewField(
		TrampolineCallContextName,
		dst.NewIdent(TrampolineCallContextType),
	)
	list.List = append([]*dst.Field{callCtx}, list.List...)
}

func (rp *RuleProcessor) buildTrampolineType(onEnter bool) *dst.FieldList {
	// Since target function parameter names might be "_", we may use the target
	// function parameters in the trampoline function, which would cause a syntax
	// error, so we assign them a specific name and use them.
	idx := 0
	renameField := func(field *dst.Field, prefix string) {
		for _, names := range field.Names {
			names.Name = fmt.Sprintf("%s%d", prefix, idx)
			idx++
		}
	}
	// Build parameter list of trampoline function
	paramList := &dst.FieldList{List: []*dst.Field{}}
	if onEnter {
		if ast.HasReceiver(rp.targetFunc) {
			recvField := dst.Clone(rp.targetFunc.Recv.List[0]).(*dst.Field)
			renameField(recvField, "recv")
			paramList.List = append(paramList.List, recvField)
		}
		for _, field := range rp.targetFunc.Type.Params.List {
			paramField := dst.Clone(field).(*dst.Field)
			renameField(paramField, "arg")
			paramList.List = append(paramList.List, paramField)
		}
	} else {
		if rp.targetFunc.Type.Results != nil {
			for _, field := range rp.targetFunc.Type.Results.List {
				retField := dst.Clone(field).(*dst.Field)
				renameField(retField, "arg")
				paramList.List = append(paramList.List, retField)
			}
		}
	}
	return paramList
}

func (rp *RuleProcessor) buildTrampolineTypes() {
	onEnterHookFunc, onExitHookFunc := rp.onEnterHookFunc, rp.onExitHookFunc
	onEnterHookFunc.Type.Params = rp.buildTrampolineType(true)
	onExitHookFunc.Type.Params = rp.buildTrampolineType(false)
	candidate := []*dst.FieldList{
		onEnterHookFunc.Type.Params,
		onExitHookFunc.Type.Params,
	}
	for _, list := range candidate {
		for i := 0; i < len(list.List); i++ {
			paramField := list.List[i]
			paramFieldType := desugarType(paramField)
			paramField.Type = ast.DereferenceOf(paramFieldType)
		}
	}
	addCallContext(onExitHookFunc.Type.Params)
}

func assignString(assignStmt *dst.AssignStmt, val string) bool {
	rhs := assignStmt.Rhs
	if len(rhs) == 1 {
		rhsExpr := rhs[0]
		if basicLit, ok2 := rhsExpr.(*dst.BasicLit); ok2 {
			if basicLit.Kind == token.STRING {
				basicLit.Value = strconv.Quote(val)
				return true
			}
		}
	}
	return false
}

func assignSliceLiteral(assignStmt *dst.AssignStmt, vals []dst.Expr) bool {
	rhs := assignStmt.Rhs
	if len(rhs) == 1 {
		rhsExpr := rhs[0]
		if compositeLit, ok := rhsExpr.(*dst.CompositeLit); ok {
			elems := compositeLit.Elts
			elems = append(elems, vals...)
			compositeLit.Elts = elems
			return true
		}
	}
	return false
}

// replenishCallContext replenishes the call context before hook invocation
func (rp *RuleProcessor) replenishCallContext(onEnter bool) bool {
	funcDecl := rp.onEnterHookFunc
	if !onEnter {
		funcDecl = rp.onExitHookFunc
	}
	for _, stmt := range funcDecl.Body.List {
		if assignStmt, ok := stmt.(*dst.AssignStmt); ok {
			lhs := assignStmt.Lhs
			if sel, ok := lhs[0].(*dst.SelectorExpr); ok {
				switch sel.Sel.Name {
				case TrampolineFuncNameIdentifier:
					util.Assert(onEnter, "sanity check")
					// callContext.FuncName = "..."
					assigned := assignString(assignStmt, rp.targetFunc.Name.Name)
					util.Assert(assigned, "sanity check")
				case TrampolinePackageNameIdentifier:
					util.Assert(onEnter, "sanity check")
					// callContext.PackageName = "..."
					assigned := assignString(assignStmt, rp.target.Name.Name)
					util.Assert(assigned, "sanity check")
				default:
					// callContext.Params = []interface{}{...} or
					// callContext.(*CallContextImpl).Params[0] = &int
					names := getNames(funcDecl.Type.Params)
					vals := make([]dst.Expr, 0, len(names))
					for i, name := range names {
						if i == 0 && !onEnter {
							// SKip first callContext parameter for after
							continue
						}
						vals = append(vals, ast.Ident(name))
					}
					assigned := assignSliceLiteral(assignStmt, vals)
					util.Assert(assigned, "sanity check")
				}
			}

		}
	}
	return true
}

// -----------------------------------------------------------------------------
// Dynamic CallContext API Generation
//
// This is somewhat challenging, as we need to generate type-aware CallContext
// APIs, which means we need to generate a bunch of switch statements to handle
// different types of parameters. Different RawFuncs in the same package may have
// different types of parameters, all of them should have their own CallContext
// implementation, thus we need to generate a bunch of CallContextImpl{suffix}
// types and methods to handle them. The suffix is generated based on the rule
// suffix, so that we can distinguish them from each other.

// implementCallContext effectively "implements" the CallContext interface by
// renaming occurrences of CallContextImpl to CallContextImpl{suffix} in the
// trampoline template
func (rp *RuleProcessor) implementCallContext(t *rules.InstFuncRule) {
	suffix := util.Crc32(t.String())
	structType := rp.callCtxDecl.Specs[0].(*dst.TypeSpec)
	util.Assert(structType.Name.Name == TrampolineCallContextImplType,
		"sanity check")
	structType.Name.Name += suffix             // type declaration
	for _, method := range rp.callCtxMethods { // method declaration
		method.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name += suffix
	}
	for _, node := range []dst.Node{rp.onEnterHookFunc, rp.onExitHookFunc} {
		dst.Inspect(node, func(node dst.Node) bool {
			if ident, ok := node.(*dst.Ident); ok {
				if ident.Name == TrampolineCallContextImplType {
					ident.Name += suffix
					return false
				}
			}
			return true
		})
	}
}

func setValue(field string, idx int, typ dst.Expr) *dst.CaseClause {
	// *(c.Params[idx].(*int)) = val.(int)
	// c.Params[idx] = val iff type is interface{}
	se := ast.SelectorExpr(ast.Ident(TrampolineCtxIdentifier), field)
	ie := ast.IndexExpr(se, ast.IntLit(idx))
	te := ast.TypeAssertExpr(ie, ast.DereferenceOf(typ))
	pe := ast.ParenExpr(te)
	de := ast.DereferenceOf(pe)
	val := ast.Ident(TrampolineValIdentifier)
	assign := ast.AssignStmt(de, ast.TypeAssertExpr(val, typ))
	if ast.IsInterfaceType(typ) {
		assign = ast.AssignStmt(ie, val)
	}
	caseClause := ast.SwitchCase(
		ast.Exprs(ast.IntLit(idx)),
		ast.Stmts(assign),
	)
	return caseClause
}

func getValue(field string, idx int, typ dst.Expr) *dst.CaseClause {
	// return *(c.Params[idx].(*int))
	// return c.Params[idx] iff type is interface{}
	se := ast.SelectorExpr(ast.Ident(TrampolineCtxIdentifier), field)
	ie := ast.IndexExpr(se, ast.IntLit(idx))
	te := ast.TypeAssertExpr(ie, ast.DereferenceOf(typ))
	pe := ast.ParenExpr(te)
	de := ast.DereferenceOf(pe)
	ret := ast.ReturnStmt(ast.Exprs(de))
	if ast.IsInterfaceType(typ) {
		ret = ast.ReturnStmt(ast.Exprs(ie))
	}
	caseClause := ast.SwitchCase(
		ast.Exprs(ast.IntLit(idx)),
		ast.Stmts(ret),
	)
	return caseClause
}

func getParamClause(idx int, typ dst.Expr) *dst.CaseClause {
	return getValue(TrampolineParamsIdentifier, idx, typ)
}

func setParamClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(TrampolineParamsIdentifier, idx, typ)
}

func getReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return getValue(TrampolineReturnValsIdentifier, idx, typ)
}

func setReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(TrampolineReturnValsIdentifier, idx, typ)
}

// desugarType desugars parameter type to its original type, if parameter
// is type of ...T, it will be converted to []T
func desugarType(param *dst.Field) dst.Expr {
	if ft, ok := param.Type.(*dst.Ellipsis); ok {
		return ast.ArrayType(ft.Elt)
	}
	return param.Type
}

func (rp *RuleProcessor) rewriteCallContext() {
	util.Assert(len(rp.callCtxMethods) > 4, "sanity check")
	var methodSetParam, methodGetParam, methodGetRetVal, methodSetRetVal *dst.FuncDecl
	for _, decl := range rp.callCtxMethods {
		switch decl.Name.Name {
		case TrampolineSetParamName:
			methodSetParam = decl
		case TrampolineGetParamName:
			methodGetParam = decl
		case TrampolineGetReturnValName:
			methodGetRetVal = decl
		case TrampolineSetReturnValName:
			methodSetRetVal = decl
		}
	}
	// Rewrite SetParam and GetParam methods
	// Don't believe what you see in template.go, we will null out it and rewrite
	// the whole switch statement
	findSwitchBlock := func(fn *dst.FuncDecl, idx int) *dst.BlockStmt {
		stmt, ok := fn.Body.List[idx].(*dst.SwitchStmt)
		util.Assert(ok, "sanity check")
		body := stmt.Body
		body.List = nil
		return body
	}
	methodSetParamBody := findSwitchBlock(methodSetParam, 1)
	methodGetParamBody := findSwitchBlock(methodGetParam, 0)
	methodSetRetValBody := findSwitchBlock(methodSetRetVal, 1)
	methodGetRetValBody := findSwitchBlock(methodGetRetVal, 0)
	idx := 0
	if ast.HasReceiver(rp.targetFunc) {
		recvType := rp.targetFunc.Recv.List[0].Type
		clause := setParamClause(idx, recvType)
		methodSetParamBody.List = append(methodSetParamBody.List, clause)
		clause = getParamClause(idx, recvType)
		methodGetParamBody.List = append(methodGetParamBody.List, clause)
		idx++
	}
	for _, param := range rp.targetFunc.Type.Params.List {
		paramType := desugarType(param)
		for range param.Names {
			clause := setParamClause(idx, paramType)
			methodSetParamBody.List =
				append(methodSetParamBody.List, clause)
			clause = getParamClause(idx, paramType)
			methodGetParamBody.List =
				append(methodGetParamBody.List, clause)
			idx++
		}
	}
	// Rewrite GetReturnVal and SetReturnVal methods
	if rp.targetFunc.Type.Results != nil {
		idx = 0
		for _, retval := range rp.targetFunc.Type.Results.List {
			retType := desugarType(retval)
			for range retval.Names {
				clause := getReturnValClause(idx, retType)
				methodGetRetValBody.List =
					append(methodGetRetValBody.List, clause)
				clause = setReturnValClause(idx, retType)
				methodSetRetValBody.List =
					append(methodSetRetValBody.List, clause)
				idx++
			}
		}
	}
}

func (rp *RuleProcessor) callHookFunc(t *rules.InstFuncRule,
	onEnter bool) error {
	traits, err := getHookParamTraits(t, onEnter)
	if err != nil {
		return err
	}
	err = rp.addHookFuncVar(t, traits, onEnter)
	if err != nil {
		return err
	}
	if onEnter {
		err = rp.callOnEnterHook(t, traits)
	} else {
		err = rp.callOnExitHook(t, traits)
	}
	if err != nil {
		return err
	}
	if !rp.replenishCallContext(onEnter) {
		return err
	}
	return nil
}

func (rp *RuleProcessor) createTrampoline(t *rules.InstFuncRule) error {
	// Materialize various declarations from template file, no one wants to see
	// a bunch of manual AST code generation, isn't it?
	err := rp.materializeTemplate()
	if err != nil {
		return err
	}
	// Implement HookContext interface methods dynamically
	rp.implementCallContext(t)
	// Rewrite type-aware HookContext APIs
	// Make all HookContext methods type-aware according to the target function
	// signature.
	rp.rewriteCallContext()
	// Rename template function to trampoline function
	rp.renameTrampolineFunc(t)
	// Build types of trampoline functions. The parameters of the Before trampoline
	// function are the same as the target function, the parameters of the After
	// trampoline function are the same as the target function.
	rp.buildTrampolineTypes()
	// Generate calls to hook functions
	if t.OnEnter != "" {
		err = rp.callHookFunc(t, true)
		if err != nil {
			return err
		}
	}
	if t.OnExit != "" {
		err = rp.callHookFunc(t, false)
		if err != nil {
			return err
		}
	}
	return nil
}
