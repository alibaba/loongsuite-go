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
	trampolineSetParamName           = "SetParam"
	trampolineGetParamName           = "GetParam"
	trampolineSetReturnValName       = "SetReturnVal"
	trampolineGetReturnValName       = "GetReturnVal"
	trampolineValIdentifier          = "val"
	trampolineCtxIdentifier          = "c"
	trampolineParamsIdentifier       = "Params"
	trampolineFuncNameIdentifier     = "FuncName"
	trampolinePackageNameIdentifier  = "PackageName"
	trampolineReturnValsIdentifier   = "ReturnVals"
	trampolineSetSkipCallName        = "SetSkipCall"
	trampolineSkipName               = "skip"
	trampolineCallContextName        = "callContext"
	trampolineCallContextType        = "CallContext"
	trampolineCallContextImplType    = "CallContextImpl"
	trampolineOnEnterName            = "OtelOnEnterTrampoline"
	trampolineOnExitName             = "OtelOnExitTrampoline"
	trampolineOnEnterNamePlaceholder = "\"OtelOnEnterNamePlaceholder\""
	trampolineOnExitNamePlaceholder  = "\"OtelOnExitNamePlaceholder\""
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
			if decl.Name.Name == trampolineOnEnterName {
				rp.onEnterHookFunc = decl
				rp.addDecl(decl)
			} else if decl.Name.Name == trampolineOnExitName {
				rp.onExitHookFunc = decl
				rp.addDecl(decl)
			} else if ast.HasReceiver(decl) {
				// We know exactly this is CallContextImpl method
				t := decl.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name
				util.Assert(t == trampolineCallContextImplType, "sanity check")
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
		util.ShouldNotReachHere()
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
	splitParams := ast.SplitMultiNameFields(target.Type.Params)
	// Find which parameter is type of interface{}
	for i, field := range splitParams.List {
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
	args := []dst.Expr{dst.NewIdent(trampolineCallContextName)}
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
	call := ast.ExprStmt(ast.CallTo(fnName, nil, args))
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
			args = append(args, dst.NewIdent(trampolineCallContextName))
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
	call := ast.ExprStmt(ast.CallTo(fnName, nil, args))
	iff := ast.IfNotNilStmt(
		dst.NewIdent(fnName),
		ast.Block(call),
		nil,
	)
	insertAtEnd(rp.onExitHookFunc, iff)
	return nil
}

// replaceTypeWithAny replaces parameter types with interface{} based on generic type parameters.
func replaceTypeWithAny(traits []ParamTrait, paramTypes, genericTypes *dst.FieldList) error {
	if len(paramTypes.List) != len(traits) {
		return ex.Newf("hook func signature can not match with target function")
	}

	for i, field := range paramTypes.List {
		trait := traits[i]
		if trait.IsInterfaceAny {
			// Hook explicitly uses interface{} for this parameter
			field.Type = ast.InterfaceType()
		} else {
			// Replace type parameters with interface{} (for linkname compatibility)
			field.Type = replaceTypeParamsWithAny(field.Type, genericTypes)
		}
	}
	return nil
}

func (rp *RuleProcessor) addHookFuncVar(t *rules.InstFuncRule,
	traits []ParamTrait, onEnter bool) error {
	paramTypes := &dst.FieldList{List: []*dst.Field{}}
	genericTypes := &dst.FieldList{List: []*dst.Field{}}
	if rp.exact {
		paramTypes, genericTypes = rp.buildTrampolineType(onEnter)
	}
	addCallContext(paramTypes)
	if rp.exact {
		// Hook functions may uses interface{} as parameter type, as some types of
		// raw function is not exposed
		err := replaceTypeWithAny(traits, paramTypes, genericTypes)
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
			if basicLit.Value == trampolineOnEnterNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnEnter)
			}
		}
		return true
	})
	rp.onExitHookFunc.Name.Name = makeName(t, rp.targetFunc, false)
	dst.Inspect(rp.onExitHookFunc, func(node dst.Node) bool {
		if basicLit, ok := node.(*dst.BasicLit); ok {
			if basicLit.Value == trampolineOnExitNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnExit)
			}
		}
		return true
	})
}

func addCallContext(list *dst.FieldList) {
	callCtx := ast.NewField(
		trampolineCallContextName,
		dst.NewIdent(trampolineCallContextType),
	)
	list.List = append([]*dst.Field{callCtx}, list.List...)
}

func (rp *RuleProcessor) buildTrampolineType(onEnter bool) (*dst.FieldList, *dst.FieldList) {
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
	// Build parameter list of trampoline function.
	// For before trampoline, it's signature is:
	// func S(h* HookContext, recv type, arg1 type, arg2 type, ...)
	// For after trampoline, it's signature is:
	// func S(h* HookContext, arg1 type, arg2 type, ...)
	// All grouped parameters (like a, b int) are expanded into separate parameters (a int, b int)
	paramTypes := &dst.FieldList{List: []*dst.Field{}}
	if onEnter {
		if ast.HasReceiver(rp.targetFunc) {
			splitRecv := ast.SplitMultiNameFields(rp.targetFunc.Recv)
			recvField := util.AssertType[*dst.Field](dst.Clone(splitRecv.List[0]))
			renameField(recvField, "recv")
			paramTypes.List = append(paramTypes.List, recvField)
		}
		splitParams := ast.SplitMultiNameFields(rp.targetFunc.Type.Params)
		for _, field := range splitParams.List {
			paramField := util.AssertType[*dst.Field](dst.Clone(field))
			renameField(paramField, "param")
			paramTypes.List = append(paramTypes.List, paramField)
		}
	} else if rp.targetFunc.Type.Results != nil {
		splitResults := ast.SplitMultiNameFields(rp.targetFunc.Type.Results)
		for _, field := range splitResults.List {
			retField := util.AssertType[*dst.Field](dst.Clone(field))
			renameField(retField, "arg")
			paramTypes.List = append(paramTypes.List, retField)
		}
	}
	// Build type parameter list of trampoline function according to the target
	// function's type parameters and receiver type parameters
	genericTypes := combineTypeParams(rp.targetFunc)
	return paramTypes, ast.CloneTypeParams(genericTypes)
}

func (rp *RuleProcessor) buildTrampolineTypes() {
	onEnterHookFunc, onExitHookFunc := rp.onEnterHookFunc, rp.onExitHookFunc
	onEnterHookFunc.Type.Params, onEnterHookFunc.Type.TypeParams = rp.buildTrampolineType(true)
	onExitHookFunc.Type.Params, onExitHookFunc.Type.TypeParams = rp.buildTrampolineType(false)
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

// populateCallContext replenishes the call context before hook invocation
func (rp *RuleProcessor) populateCallContext(onEnter bool) bool {
	funcDecl := rp.onEnterHookFunc
	if !onEnter {
		funcDecl = rp.onExitHookFunc
	}
	for _, stmt := range funcDecl.Body.List {
		if assignStmt, ok := stmt.(*dst.AssignStmt); ok {
			lhs := assignStmt.Lhs
			if sel, ok := lhs[0].(*dst.SelectorExpr); ok {
				switch sel.Sel.Name {
				case trampolineFuncNameIdentifier:
					util.Assert(onEnter, "sanity check")
					// callContext.FuncName = "..."
					assigned := assignString(assignStmt, rp.targetFunc.Name.Name)
					util.Assert(assigned, "sanity check")
				case trampolinePackageNameIdentifier:
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
	util.Assert(structType.Name.Name == trampolineCallContextImplType,
		"sanity check")
	structType.Name.Name += suffix             // type declaration
	for _, method := range rp.callCtxMethods { // method declaration
		method.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name += suffix
	}
	for _, node := range []dst.Node{rp.onEnterHookFunc, rp.onExitHookFunc} {
		dst.Inspect(node, func(node dst.Node) bool {
			if ident, ok := node.(*dst.Ident); ok {
				if ident.Name == trampolineCallContextImplType {
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
	se := ast.SelectorExpr(ast.Ident(trampolineCtxIdentifier), field)
	ie := ast.IndexExpr(se, ast.IntLit(idx))
	te := ast.TypeAssertExpr(ie, ast.DereferenceOf(typ))
	pe := ast.ParenExpr(te)
	de := ast.DereferenceOf(pe)
	val := ast.Ident(trampolineValIdentifier)
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
	se := ast.SelectorExpr(ast.Ident(trampolineCtxIdentifier), field)
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
	return getValue(trampolineParamsIdentifier, idx, typ)
}

func setParamClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(trampolineParamsIdentifier, idx, typ)
}

func getReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return getValue(trampolineReturnValsIdentifier, idx, typ)
}

func setReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(trampolineReturnValsIdentifier, idx, typ)
}

// extractReceiverTypeParams extracts type parameters from a receiver type expression
// For example: *GenStruct[T] or GenStruct[T, U] -> FieldList with T and U as type parameters
func extractReceiverTypeParams(recvType dst.Expr) *dst.FieldList {
	switch t := recvType.(type) {
	case *dst.StarExpr:
		// *GenStruct[T] - recurse into X
		return extractReceiverTypeParams(t.X)
	case *dst.IndexExpr:
		// GenStruct[T] - single type parameter
		if ident, ok := t.Index.(*dst.Ident); ok {
			return &dst.FieldList{
				List: []*dst.Field{{
					Names: []*dst.Ident{ident},
					Type:  ast.Ident("any"), // Type constraint for the parameter
				}},
			}
		}
	case *dst.IndexListExpr:
		// GenStruct[T, U, ...] - multiple type parameters
		fields := make([]*dst.Field, 0, len(t.Indices))
		for _, idx := range t.Indices {
			if ident, ok := idx.(*dst.Ident); ok {
				fields = append(fields, &dst.Field{
					Names: []*dst.Ident{ident},
					Type:  ast.Ident("any"), // Type constraint for the parameter
				})
			}
		}
		if len(fields) > 0 {
			return &dst.FieldList{List: fields}
		}
	}
	return nil
}

// combineTypeParams combines type parameters from the receiver and function type parameters.
// For methods on generic types, it extracts type parameters from the receiver and merges
// them with the function's type parameters.
// Receiver type parameters come first, followed by function type parameters.
//
// Example:
//
//	Original: func (c *Container[K]) Transform[V any]() V
//	Result: [K, V]
//
//	Generated trampolines:
//	  func OtelBeforeTrampoline_Container_Transform[K comparable, V any](
//	      hookContext *HookContext,
//	      recv0 *Container[K],  // ← Uses K
//	  ) { ... }
//
//	  func OtelAfterTrampoline_Container_Transform[K comparable, V any](
//	      hookContext *HookContext,
//	      arg0 *V,  // ← Uses V (return type)
//	  ) { ... }
func combineTypeParams(targetFunc *dst.FuncDecl) *dst.FieldList {
	var trampolineTypeParams *dst.FieldList
	if ast.HasReceiver(targetFunc) {
		receiverTypeParams := extractReceiverTypeParams(targetFunc.Recv.List[0].Type)
		if receiverTypeParams != nil {
			trampolineTypeParams = receiverTypeParams
		}
	}
	if targetFunc.Type.TypeParams != nil {
		if trampolineTypeParams == nil {
			trampolineTypeParams = targetFunc.Type.TypeParams
		} else {
			combined := &dst.FieldList{List: make([]*dst.Field, 0)}
			combined.List = append(combined.List, trampolineTypeParams.List...)
			combined.List = append(combined.List, targetFunc.Type.TypeParams.List...)
			trampolineTypeParams = combined
		}
	}
	return trampolineTypeParams
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
		case trampolineSetParamName:
			methodSetParam = decl
		case trampolineGetParamName:
			methodGetParam = decl
		case trampolineGetReturnValName:
			methodGetRetVal = decl
		case trampolineSetReturnValName:
			methodSetRetVal = decl
		}
	}
	combinedTypeParams := combineTypeParams(rp.targetFunc)

	// Rewrite SetParam and GetParam methods
	// Don't believe what you see in template, we will null out it and rewrite
	// the whole switch statement
	findSwitchBlock := func(fn *dst.FuncDecl, idx int) *dst.BlockStmt {
		stmt := util.AssertType[*dst.SwitchStmt](fn.Body.List[idx])
		body := stmt.Body
		body.List = nil
		return body
	}

	// For generic functions, SetParam and SetReturnVal should panic
	// as modifying parameters/return values is unsupported for generic functions
	if combinedTypeParams != nil {
		makeMethodPanic(methodSetParam, "SetParam is unsupported for generic functions")
		makeMethodPanic(methodSetRetVal, "SetReturnVal is unsupported for generic functions")
		methodGetParamBody := findSwitchBlock(methodGetParam, 0)
		methodGetRetValBody := findSwitchBlock(methodGetRetVal, 0)

		rp.rewriteHookContextParams(nil, methodGetParamBody, combinedTypeParams)
		rp.rewriteHookContextResults(nil, methodGetRetValBody, combinedTypeParams)
	} else {
		methodSetParamBody := findSwitchBlock(methodSetParam, 1)
		methodGetParamBody := findSwitchBlock(methodGetParam, 0)
		methodSetRetValBody := findSwitchBlock(methodSetRetVal, 1)
		methodGetRetValBody := findSwitchBlock(methodGetRetVal, 0)

		rp.rewriteHookContextParams(methodSetParamBody, methodGetParamBody, combinedTypeParams)
		rp.rewriteHookContextResults(methodSetRetValBody, methodGetRetValBody, combinedTypeParams)
	}
}

func (rp *RuleProcessor) rewriteHookContextParams(
	methodSetParamBody, methodGetParamBody *dst.BlockStmt,
	combinedTypeParams *dst.FieldList,
) {
	isGeneric := combinedTypeParams != nil
	idx := 0
	if ast.HasReceiver(rp.targetFunc) {
		splitRecv := ast.SplitMultiNameFields(rp.targetFunc.Recv)
		recvType := replaceTypeParamsWithAny(splitRecv.List[0].Type, combinedTypeParams)
		if !isGeneric {
			clause := setParamClause(idx, recvType)
			methodSetParamBody.List = append(methodSetParamBody.List, clause)
		}
		clause := getParamClause(idx, recvType)
		methodGetParamBody.List = append(methodGetParamBody.List, clause)
		idx++
	}
	splitParams := ast.SplitMultiNameFields(rp.targetFunc.Type.Params)
	for _, param := range splitParams.List {
		paramType := replaceTypeParamsWithAny(desugarType(param), combinedTypeParams)
		if !isGeneric {
			clause := setParamClause(idx, paramType)
			methodSetParamBody.List = append(methodSetParamBody.List, clause)
		}
		clause := getParamClause(idx, paramType)
		methodGetParamBody.List = append(methodGetParamBody.List, clause)
		idx++
	}
}

func (rp *RuleProcessor) rewriteHookContextResults(
	methodSetRetValBody, methodGetRetValBody *dst.BlockStmt,
	combinedTypeParams *dst.FieldList,
) {
	isGeneric := combinedTypeParams != nil
	if rp.targetFunc.Type.Results != nil {
		idx := 0
		splitResults := ast.SplitMultiNameFields(rp.targetFunc.Type.Results)
		for _, retval := range splitResults.List {
			retType := replaceTypeParamsWithAny(desugarType(retval), combinedTypeParams)
			clause := getReturnValClause(idx, retType)
			methodGetRetValBody.List = append(methodGetRetValBody.List, clause)
			if !isGeneric {
				clause = setReturnValClause(idx, retType)
				methodSetRetValBody.List = append(methodSetRetValBody.List, clause)
			}
			idx++
		}
	}
}

// makeMethodPanic replaces a method's body with a panic statement
func makeMethodPanic(method *dst.FuncDecl, message string) {
	panicStmt := ast.ExprStmt(
		ast.CallTo("panic", nil, []dst.Expr{
			&dst.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(message),
			},
		}),
	)
	method.Body.List = []dst.Stmt{panicStmt}
}

// isTypeParameter checks if a type expression is a bare type parameter identifier
func isTypeParameter(t dst.Expr, typeParams *dst.FieldList) bool {
	if typeParams == nil {
		return false
	}
	ident, ok := t.(*dst.Ident)
	if !ok {
		return false
	}
	// Check if this identifier matches any type parameter name
	for _, field := range typeParams.List {
		for _, name := range field.Names {
			if name.Name == ident.Name {
				return true
			}
		}
	}
	return false
}

// replaceTypeParamsWithAny replaces type parameters with interface{} for use in
// non-generic contexts like HookContextImpl methods
func replaceTypeParamsWithAny(t dst.Expr, typeParams *dst.FieldList) dst.Expr {
	if isTypeParameter(t, typeParams) {
		return ast.InterfaceType()
	}

	// For complex types like *T, []T, map[K]V, etc., handle them recursively
	switch tType := t.(type) {
	case *dst.StarExpr:
		// *T -> *interface{}
		return ast.DereferenceOf(replaceTypeParamsWithAny(tType.X, typeParams))
	case *dst.ArrayType:
		// []T -> []interface{}
		return ast.ArrayType(replaceTypeParamsWithAny(tType.Elt, typeParams))
	case *dst.MapType:
		// map[K]V -> map[interface{}]interface{}
		return &dst.MapType{
			Key:   replaceTypeParamsWithAny(tType.Key, typeParams),
			Value: replaceTypeParamsWithAny(tType.Value, typeParams),
		}
	case *dst.ChanType:
		// chan T, <-chan T, chan<- T -> chan interface{}, etc.
		return &dst.ChanType{
			Dir:   tType.Dir,
			Value: replaceTypeParamsWithAny(tType.Value, typeParams),
		}
	case *dst.IndexExpr:
		// GenStruct[T] -> interface{} (for generic receiver methods)
		// The hook function expects interface{} for generic types
		return ast.InterfaceType()
	case *dst.IndexListExpr:
		// GenStruct[T, U] -> interface{} (for generic receiver methods with multiple type params)
		return ast.InterfaceType()
	case *dst.Ident, *dst.SelectorExpr, *dst.InterfaceType:
		// Base types without type parameters, return as-is
		return t
	case *dst.Ellipsis:
		// ...T -> ...interface{}
		return ast.Ellipsis(replaceTypeParamsWithAny(tType.Elt, typeParams))
	default:
		// Unsupported cases:
		// - *dst.FuncType (function types with type parameters)
		// - Other uncommon type expressions
		// util.Unimplemented(fmt.Sprintf("unexpected generic type: %T", tType))
		return t
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
	if !rp.populateCallContext(onEnter) {
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
