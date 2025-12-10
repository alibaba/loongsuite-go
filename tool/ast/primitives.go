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

package ast

import (
	"fmt"
	"go/token"

	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

const (
	IdentNil    = "nil"
	IdentTrue   = "true"
	IdentFalse  = "false"
	IdentIgnore = "_"
)

// -----------------------------------------------------------------------------
// AST Primitives
//
// This file provides essential primitives for AST manipulation, including common
// identifier constants, type checking, expression and so on.
//
// The primitives defined here serve as building blocks for higher-level AST
// operations throughout the instrumentation toolchain, ensuring consistent
// handling of common AST patterns and reducing code duplication.

func AddressOf(name string) *dst.UnaryExpr {
	return &dst.UnaryExpr{Op: token.AND, X: Ident(name)}
}

// CallTo creates a call expression to a function with optional type arguments for generics.
// For non-generic functions (typeArgs is nil or empty), creates a simple call: Foo(args...)
// For generic functions with type arguments, creates: Foo[T1, T2](args...)
func CallTo(name string, typeArgs *dst.FieldList, args []dst.Expr) *dst.CallExpr {
	if typeArgs == nil || len(typeArgs.List) == 0 {
		return &dst.CallExpr{
			Fun:  &dst.Ident{Name: name},
			Args: args,
		}
	}

	var indices []dst.Expr
	for _, field := range typeArgs.List {
		for _, ident := range field.Names {
			indices = append(indices, Ident(ident.Name))
		}
	}
	var fun dst.Expr
	if len(indices) == 1 {
		fun = IndexExpr(Ident(name), indices[0])
	} else {
		fun = IndexListExpr(Ident(name), indices)
	}
	return &dst.CallExpr{
		Fun:  fun,
		Args: args,
	}
}
func Ident(name string) *dst.Ident {
	return &dst.Ident{
		Name: name,
	}
}

func Nil() dst.Expr {
	return Ident(IdentNil)
}

func StringLit(value string) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.STRING,
		Value: fmt.Sprintf("%q", value),
	}
}

func IntLit(value int) *dst.BasicLit {
	return &dst.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", value),
	}
}

func Block(stmt dst.Stmt) *dst.BlockStmt {
	return &dst.BlockStmt{
		List: []dst.Stmt{
			stmt,
		},
	}
}

func BlockStmts(stmts ...dst.Stmt) *dst.BlockStmt {
	return &dst.BlockStmt{
		List: stmts,
	}
}

func Exprs(exprs ...dst.Expr) []dst.Expr {
	return exprs
}

func Stmts(stmts ...dst.Stmt) []dst.Stmt {
	return stmts
}

func SelectorExpr(x dst.Expr, sel string) *dst.SelectorExpr {
	return &dst.SelectorExpr{
		X:   dst.Clone(x).(dst.Expr),
		Sel: Ident(sel),
	}
}

func Ellipsis(elt dst.Expr) *dst.Ellipsis {
	return &dst.Ellipsis{
		Elt: elt,
	}
}

func IndexExpr(x dst.Expr, index dst.Expr) *dst.IndexExpr {
	return &dst.IndexExpr{
		X:     dst.Clone(x).(dst.Expr),
		Index: dst.Clone(index).(dst.Expr),
	}
}

func IndexListExpr(x dst.Expr, indices []dst.Expr) *dst.IndexListExpr {
	e := util.AssertType[dst.Expr](dst.Clone(x))
	return &dst.IndexListExpr{
		X:       e,
		Indices: indices,
	}
}

func TypeAssertExpr(x dst.Expr, typ dst.Expr) *dst.TypeAssertExpr {
	return &dst.TypeAssertExpr{
		X:    x,
		Type: dst.Clone(typ).(dst.Expr),
	}
}

func ParenExpr(x dst.Expr) *dst.ParenExpr {
	return &dst.ParenExpr{
		X: dst.Clone(x).(dst.Expr),
	}
}

func NewField(name string, typ dst.Expr) *dst.Field {
	newField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  typ,
	}
	return newField
}

func BoolTrue() *dst.BasicLit {
	return &dst.BasicLit{Value: IdentTrue}
}

func BoolFalse() *dst.BasicLit {
	return &dst.BasicLit{Value: IdentFalse}
}

func InterfaceType() *dst.InterfaceType {
	return &dst.InterfaceType{
		Methods: &dst.FieldList{Opening: true, Closing: true},
	}
}

func ArrayType(elem dst.Expr) *dst.ArrayType {
	return &dst.ArrayType{Elt: elem}
}

func IfStmt(init dst.Stmt, cond dst.Expr,
	body, elseBody *dst.BlockStmt) *dst.IfStmt {
	return &dst.IfStmt{
		Init: dst.Clone(init).(dst.Stmt),
		Cond: dst.Clone(cond).(dst.Expr),
		Body: dst.Clone(body).(*dst.BlockStmt),
		Else: dst.Clone(elseBody).(*dst.BlockStmt),
	}
}

func IfNotNilStmt(cond dst.Expr, body, elseBody *dst.BlockStmt) *dst.IfStmt {
	var elseB dst.Stmt
	if elseBody == nil {
		elseB = nil
	} else {
		elseB = dst.Clone(elseBody).(dst.Stmt)
	}
	return &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.Clone(cond).(dst.Expr),
			Op: token.NEQ,
			Y:  &dst.Ident{Name: IdentNil},
		},
		Body: dst.Clone(body).(*dst.BlockStmt),
		Else: elseB,
	}
}

func EmptyStmt() *dst.EmptyStmt {
	return &dst.EmptyStmt{}
}

func ExprStmt(expr dst.Expr) *dst.ExprStmt {
	return &dst.ExprStmt{X: dst.Clone(expr).(dst.Expr)}
}

func DeferStmt(call *dst.CallExpr) *dst.DeferStmt {
	return &dst.DeferStmt{Call: dst.Clone(call).(*dst.CallExpr)}
}

func ReturnStmt(results []dst.Expr) *dst.ReturnStmt {
	return &dst.ReturnStmt{Results: results}
}

func AssignStmt(lhs, rhs dst.Expr) *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{lhs},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{rhs},
	}
}

func DefineStmts(lhs, rhs []dst.Expr) *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: lhs,
		Tok: token.DEFINE,
		Rhs: rhs,
	}
}

func SwitchCase(list []dst.Expr, stmts []dst.Stmt) *dst.CaseClause {
	return &dst.CaseClause{
		List: list,
		Body: stmts,
	}
}

func NewVarDecl(name string, paramTypes *dst.FieldList) *dst.GenDecl {
	return &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{
					{Name: name},
				},
				Type: &dst.FuncType{
					Func:   false,
					Params: paramTypes,
				},
			},
		},
	}
}

func DereferenceOf(expr dst.Expr) dst.Expr {
	return &dst.StarExpr{X: expr}
}

func KeyValueExpr(key string, value dst.Expr) *dst.KeyValueExpr {
	return &dst.KeyValueExpr{
		Key:   Ident(key),
		Value: value,
	}
}

func CompositeLit(t dst.Expr, elts []dst.Expr) *dst.CompositeLit {
	return &dst.CompositeLit{
		Type: t,
		Elts: elts,
	}
}

func StructLit(typeName string, fields ...*dst.KeyValueExpr) dst.Expr {
	exprs := make([]dst.Expr, len(fields))
	for i, field := range fields {
		exprs[i] = field
	}
	return &dst.UnaryExpr{
		Op: token.AND,
		X:  CompositeLit(Ident(typeName), exprs),
	}
}
