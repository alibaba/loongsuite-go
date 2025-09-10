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

package ast

import (
	"fmt"
	"go/token"
	"regexp"

	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

func MakeUnusedIdent(ident *dst.Ident) *dst.Ident {
	ident.Name = IdentIgnore
	return ident
}

func IsUnusedIdent(ident *dst.Ident) bool {
	return ident.Name == IdentIgnore
}

func IsStringLit(expr dst.Expr, val string) bool {
	lit, ok := expr.(*dst.BasicLit)
	return ok &&
		lit.Kind == token.STRING &&
		lit.Value == fmt.Sprintf("%q", val)
}

func IsInterfaceType(typ dst.Expr) bool {
	_, ok := typ.(*dst.InterfaceType)
	return ok
}

func IsEllipsis(typ dst.Expr) bool {
	_, ok := typ.(*dst.Ellipsis)
	return ok
}

func AddStructField(decl dst.Decl, name string, typ string) {
	gen, ok := decl.(*dst.GenDecl)
	util.Assert(ok, "decl is not a GenDecl")
	fd := NewField(name, Ident(typ))
	st := gen.Specs[0].(*dst.TypeSpec).Type.(*dst.StructType)
	st.Fields.List = append(st.Fields.List, fd)
}

func addImport(root *dst.File, paths ...string) *dst.GenDecl {
	importStmt := &dst.GenDecl{Tok: token.IMPORT}
	specs := make([]dst.Spec, 0)
	for _, path := range paths {
		spec := &dst.ImportSpec{
			Path: &dst.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("%q", path),
			},
			Name: &dst.Ident{Name: IdentIgnore},
		}
		specs = append(specs, spec)
	}
	importStmt.Specs = specs
	root.Decls = append([]dst.Decl{importStmt}, root.Decls...)
	return importStmt
}

func AddImportForcely(root *dst.File, paths ...string) *dst.GenDecl {
	return addImport(root, paths...)
}

func RemoveImport(root *dst.File, path string) *dst.ImportSpec {
	for j, decl := range root.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok &&
			genDecl.Tok == token.IMPORT {
			for i, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*dst.ImportSpec); ok {
					if importSpec.Path.Value == fmt.Sprintf("%q", path) {
						genDecl.Specs =
							append(genDecl.Specs[:i],
								genDecl.Specs[i+1:]...)
						if len(genDecl.Specs) == 0 {
							root.Decls =
								append(root.Decls[:j], root.Decls[j+1:]...)
						}
						return importSpec
					}
				}
			}
		}
	}
	return nil
}

func FindImport(root *dst.File, path string) *dst.ImportSpec {
	for _, decl := range root.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok &&
			genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*dst.ImportSpec); ok {
					if importSpec.Path.Value == fmt.Sprintf("%q", path) {
						return importSpec
					}
				}
			}
		}
	}
	return nil
}

func HasReceiver(fn *dst.FuncDecl) bool {
	return fn.Recv != nil && len(fn.Recv.List) > 0
}

func FindFuncDecl(root *dst.File, name string) *dst.FuncDecl {
	for _, decl := range root.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func isValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

func MatchFuncDecl(decl dst.Decl, function string, receiverType string) bool {
	util.Assert(isValidRegex(function), "invalid function name pattern")

	funcDecl, ok := decl.(*dst.FuncDecl)
	if !ok {
		return false
	}
	re := regexp.MustCompile("^" + function + "$") // strict match
	if !re.MatchString(funcDecl.Name.Name) {
		return false
	}
	if receiverType != "" {
		re = regexp.MustCompile("^" + receiverType + "$") // strict match
		if !HasReceiver(funcDecl) {
			return re.MatchString("")
		}
		switch recvTypeExpr := funcDecl.Recv.List[0].Type.(type) {
		case *dst.StarExpr:
			if _, ok := recvTypeExpr.X.(*dst.Ident); !ok {
				// This is a generic type, we don't support it yet
				return false
			}
			t := "*" + recvTypeExpr.X.(*dst.Ident).Name
			return re.MatchString(t)
		case *dst.Ident:
			t := recvTypeExpr.Name
			return re.MatchString(t)
		case *dst.IndexExpr:
			// This is a generic type, we don't support it yet
			return false
		default:
			msg := fmt.Sprintf("unexpected receiver type: %T", recvTypeExpr)
			util.UnimplementedT(msg)
		}
	} else {
		if HasReceiver(funcDecl) {
			return false
		}
	}
	return true
}

func MatchStructDecl(decl dst.Decl, structType string) bool {
	if genDecl, ok := decl.(*dst.GenDecl); ok {
		if genDecl.Tok == token.TYPE {
			if typeSpec, ok := genDecl.Specs[0].(*dst.TypeSpec); ok {
				if typeSpec.Name.Name == structType {
					return true
				}
			}
		}
	}
	return false
}
