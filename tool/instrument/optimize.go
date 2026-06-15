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
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// == Code conjured by yyang, Jun. 2024 ==

// -----------------------------------------------------------------------------
// Trampoline Optimization
//
// Since trampoline-jump-if and trampoline functions are performance-critical,
// we are trying to optimize them as much as possible. The standard form of
// trampoline-jump-if looks like
//
//	if ctx, skip := otel_trampoline_onenter(&arg); skip {
//	    otel_trampoline_onexit(ctx, &retval)
//	    return ...
//	} else {
//	    defer otel_trampoline_onexit(ctx, &retval)
//	    ...
//	}
//
// The obvious optimization opportunities are cases when onEnter or onExit hooks
// are not present. For the latter case, we can replace the defer statement to
// empty statement, you might argue that we can remove the whole else block, but
// there might be more than one trampoline-jump-if in the same function, they are
// nested in the else block, i.e.
//
//	if ctx, skip := otel_trampoline_onenter(&arg); skip {
//	    otel_trampoline_onexit(ctx, &retval)
//	    return ...
//	} else {
//	    ;
//	    ...
//	}
//
// For the former case, it's a bit more complicated. We need to manually construct
// CallContext on the fly and pass it to onExit trampoline defer call and rewrite
// the whole condition to always false. The corresponding code snippet is
//
//	if false {
//	    ;
//	} else {
//	    defer otel_trampoline_onexit(&CallContext{...}, &retval)
//	    ...
//	}
//
// The if skeleton should be kept as is, otherwise inlining of trampoline-jump-if
// will not work. During compiling, the dce and sccp passes will remove the whole
// then block. That's not the whole story. We can further optimize the tjump iff
// the onEnter hook does not use SkipCall. In this case, we can rewrite condition
// of trampoline-jump-if to always false, remove return statement in then block,
// they are memory-aware and may generate memory SSA values during compilation.
//
//	if ctx,_ := otel_trampoline_onenter(&arg); false {
//	    ;
//	} else {
//	    defer otel_trampoline_onexit(ctx, &retval)
//	    ...
//	}
//
// The compiler responsible for hoisting the initialization statement out of the
// if skeleton, and the dce and sccp passes will remove the whole then block. All
// these trampoline functions looks as if they are executed sequentially, i.e.
//
//	ctx,_ := otel_trampoline_onenter(&arg);
//	defer otel_trampoline_onexit(ctx, &retval)
//
// Note that this optimization pass is fraigle as it really heavily depends on
// the structure of trampoline-jump-if and trampoline functions. Any change in
// tjump should be carefully examined.

// TJump describes a trampoline-jump-if optimization candidate
type TJump struct {
	target *dst.FuncDecl       // Target function we are hooking on
	ifStmt *dst.IfStmt         // Trampoline-jump-if statement
	rule   *rules.InstFuncRule // Rule associated with the trampoline-jump-if
}

func mustTJump(ifStmt *dst.IfStmt) {
	util.Assert(len(ifStmt.Decs.If) == 1, "must be a trampoline-jump-if")
	desc := ifStmt.Decs.If[0]
	util.Assert(desc == TJumpLabel, "must be a trampoline-jump-if")
}

func (rp *RuleProcessor) removeOnExitTrampolineCall(tjump *TJump) error {
	ifStmt := tjump.ifStmt
	elseBlock := ifStmt.Else.(*dst.BlockStmt)
	for i, stmt := range elseBlock.List {
		if _, ok := stmt.(*dst.DeferStmt); ok {
			// Replace defer statement with an empty statement
			elseBlock.List[i] = ast.EmptyStmt()
			if config.GetConf().Verbose {
				util.Log("Optimize tjump branch in %s",
					tjump.target.Name.Name)
			}
			break
		} else if _, ok := stmt.(*dst.IfStmt); ok {
			// Expected statement type and do nothing
		} else {
			// Unexpected statement type
			util.ShouldNotReachHere()
		}
	}
	return nil
}

// newCallContextImpl constructs a new CallContextImpl structure literal and
// replenishes its Params && ReturnValues field with addresses of all arguments.
// The CallContextImpl structure is used to pass arguments to the exit trampoline
func newCallContextImpl(tjump *TJump) dst.Expr {
	targetFunc := tjump.target
	structName := trampolineCallContextImplType + util.Crc32(tjump.rule.String())

	// Build params slice: []interface{}{&param1, &param2, ...}
	// Use createHookArgs to handle underscore parameters correctly
	paramNames := collectArguments(targetFunc)
	paramExprs := createHookArgs(paramNames)
	paramsSlice := ast.CompositeLit(
		ast.ArrayType(ast.InterfaceType()),
		paramExprs,
	)

	// Build returnVals slice: []interface{}{&retval1, &retval2, ...}
	returnExprs := make([]dst.Expr, 0)
	if targetFunc.Type.Results != nil {
		returnNames := collectReturnValues(targetFunc)
		returnExprs = createHookArgs(returnNames)
	}
	returnValsSlice := ast.CompositeLit(
		ast.ArrayType(ast.InterfaceType()),
		returnExprs,
	)

	// Build the struct literal: &HookContextImpl{params:..., returnVals:...}
	return ast.StructLit(
		structName,
		ast.KeyValueExpr(trampolineParamsIdentifier, paramsSlice),
		ast.KeyValueExpr(trampolineReturnValsIdentifier, returnValsSlice),
	)
}

func (rp *RuleProcessor) removeOnEnterTrampolineCall(tjump *TJump) error {
	// Construct CallContext on the fly and pass to onExit trampoline defer call
	callContextExpr := newCallContextImpl(tjump)
	// Find defer call to onExit and replace its call context with new one
	found := false
	for _, stmt := range tjump.ifStmt.Else.(*dst.BlockStmt).List {
		// Replace call context argument of defer statement to structure literal
		if deferStmt, ok := stmt.(*dst.DeferStmt); ok {
			args := deferStmt.Call.Args
			util.Assert(len(args) >= 1, "must have at least one argument")
			args[0] = callContextExpr
			found = true
			break
		}
	}
	util.Assert(found, "defer statement not found")
	// Rewrite condition of trampoline-jump-if to always false and null out its
	// initialization statement and then block
	tjump.ifStmt.Init = nil
	tjump.ifStmt.Cond = ast.BoolFalse()
	tjump.ifStmt.Body = ast.Block(ast.EmptyStmt())
	if config.GetConf().Verbose {
		util.Log("Optimize tjump branch in %s", tjump.target.Name.Name)
	}
	// Remove generated onEnter trampoline function
	removed := rp.removeDeclWhen(func(d dst.Decl) bool {
		if funcDecl, ok := d.(*dst.FuncDecl); ok {
			return funcDecl.Name.Name == makeName(tjump.rule, tjump.target, true)
		}
		return false
	})
	if removed == nil {
		return ex.Newf("onEnter trampoline not found")
	}
	return nil
}

// canFlattenTJump checks if the tjump can be safely flattened based on
// the hook function's usage of HookContext. Returns true if:
// 1. SetSkipCall is never called (so skip is always false)
// 2. The HookContext parameter is only used as a receiver for method calls
func canFlattenTJump(hookFunc *dst.FuncDecl) bool {
	// Check if the hook function contains any "SetSkipCall" string
	// If found, the trampoline-jump-if cannot be flattened
	found := false
	dst.Inspect(hookFunc, func(node dst.Node) bool {
		if ident, ok := node.(*dst.Ident); ok {
			if strings.Contains(ident.Name, trampolineSetSkipCallName) {
				found = true
				return false
			}
		}
		if found {
			return false
		}
		return true
	})
	if found {
		return false
	}

	// Check if the hook context parameter escapes (used for non-method calls)
	escape := false
	hookContextParam := hookFunc.Type.Params.List[0].Names[0].Name
	if hookContextParam == ast.IdentIgnore {
		// If the parameter is ignored, it doesn't escape because it is not used
		return true
	}
	dst.Inspect(hookFunc.Body, func(n dst.Node) bool {
		if escape {
			return false
		}
		switch n := n.(type) {
		case *dst.SelectorExpr:
			// Check if ictx is used as method receiver: ictx.Method()
			if id, ok := n.X.(*dst.Ident); ok && id.Name == hookContextParam {
				// Valid usage.
				// Return false to stop visiting children, so we don't visit
				// the Ident "ictx", which would be caught by the case below.
				return false
			}
		case *dst.Ident:
			// If we encounter ictx here, it means it wasn't part of a method
			// call receiver (because we returned false above). So it is an
			// invalid usage.
			if n.Name == hookContextParam {
				escape = true
				return false
			}
		}
		return true
	})
	return !escape
}

// flattenTJump transforms the trampoline-jump-if AST to a flattened form.
// It sets the condition to false and empties the then block, effectively
// converting the branching pattern to sequential execution.
func flattenTJump(tjump *TJump, removedOnExit bool) error {
	// The current standard tjump pattern is as follows:
	//
	// 	if ctx, skip := otel_trampoline_before(&arg); skip {
	// 		otel_trampoline_after(ctx, &retval)
	// 		return ...
	// 	} else {
	// 		defer otel_trampoline_after(ctx, &retval)
	// 		...
	// 	}
	//
	// A key optimization opportunity lies in "skip", which is highly likely to
	// be false. In this scenario, tjump can be optimized into the following form:
	//
	// 	ctx,_ := otel_trampoline_before(&arg);
	// 	defer otel_trampoline_after(ctx, &retval)
	//
	// Consider the following hook code
	//
	//	func hookFunc(ictx HookContext, arg1....) {
	// 		ictx.SetSkipCall()
	// 		passTo(ictx)
	// 		var escape interface{} = ictx
	//	}
	//
	// This optimization can be applied when the HookContext parameter meets these conditions:
	// 1. SetSkipCall is never called (so skip is always false)
	// 2. The HookContext is ONLY used as a receiver for method calls
	//
	// Allowed usage (optimization can proceed):
	//	func hookFunc(ictx HookContext, arg1....) {
	// 		ictx.GetParam(0)      // ✓ Method calls are allowed
	// 		ictx.SetParam(1, val) // ✓ Method calls are allowed
	//	}
	//
	// Disallowed usage (prevents optimization):
	//	func hookFunc(ictx HookContext, arg1....) {
	// 		ictx.SetSkipCall()           // ✗ SetSkipCall prevents optimization
	// 		passTo(ictx)                 // ✗ Cannot pass as argument
	// 		var escape interface{} = ictx // ✗ Cannot assign to variable
	// 		_ = ictx                     // ✗ Cannot use in any assignment
	//	}
	//
	// When both conditions are met, the HookContext doesn't escape and we can
	// safely flatten the trampoline-jump-if pattern, significantly boosting performance.
	// 1. If "SetSkipCall" string never appears in the Hook code,
	// 2. and HookContext parameter is never used for purposes other than method
	// calls (e.g. assignment, pass as args, etc.), then the HookContext parameter
	// does not escape, and the tjump represents a valid candidate for optimization.
	// This would significantly boost performance.
	ifStmt := tjump.ifStmt
	initStmt := util.AssertType[*dst.AssignStmt](ifStmt.Init)
	util.Assert(len(initStmt.Lhs) == 2, "must have two lhs")

	// Set condition to false and empty the then block
	ifStmt.Cond = ast.BoolFalse()
	ifStmt.Body = ast.Block(ast.EmptyStmt())

	if removedOnExit {
		// We removed the last reference to hook context after nulling out body
		// block, at this point, all lhs are unused, replace assignment to simple
		// function call
		ifStmt.Init = ast.ExprStmt(initStmt.Rhs[0])
		// TODO: Remove After declaration as well
	} else {
		// Otherwise, mark skipCall identifier as unused
		skipCallIdent := util.AssertType[*dst.Ident](initStmt.Lhs[1])
		ast.MakeUnusedIdent(skipCallIdent)
	}
	return nil
}

func stripTJumpLabel(tjump *TJump) {
	ifStmt := tjump.ifStmt
	ifStmt.Decs.If = nil
}

func (rp *RuleProcessor) optimizeTJumps() (err error) {
	for _, tjump := range rp.trampolineJumps {
		mustTJump(tjump.ifStmt)
		// Strip the trampoline-jump-if anchor label as no longer needed
		stripTJumpLabel(tjump)

		// No onExit hook present? Simply remove defer call to onExit trampoline.
		// Why we don't remove the whole else block of trampoline-jump-if? Well,
		// because there might be more than one trampoline-jump-if in the same
		// function, they are nested in the else block. See findJumpPoint for
		// more details.
		// TODO: Remove corresponding CallContextImpl methods
		rule := tjump.rule
		removedOnExit := false
		if rule.OnExit == "" {
			err = rp.removeOnExitTrampolineCall(tjump)
			if err != nil {
				return err
			}
			removedOnExit = true
		}

		// No onEnter hook present? Construct CallContext on the fly and pass it
		// to onExit trampoline defer call and rewrite the whole condition to
		// always false, then null out its initialization statement.
		if rule.OnEnter == "" {
			err = rp.removeOnEnterTrampolineCall(tjump)
			if err != nil {
				return err
			}
		}

		// No SkipCall used in onEnter hook? Rewrite cond of trampoline-jump-if
		// to always false, and remove return statement in then block, they are
		// memory aware and may generate memory SSA values during compilation.
		// This further simplifies the trampoline-jump-if and gives more chances
		// for optimization passes to kick in.
		if rule.OnEnter != "" {
			hookFunc, err := getHookFunc(tjump.rule, true)
			if err != nil {
				return err
			}
			if canFlattenTJump(hookFunc) {
				err1 := flattenTJump(tjump, removedOnExit)
				if err1 != nil {
					return err1
				}
			}
		}
	}
	return nil
}
