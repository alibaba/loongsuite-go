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

// -----------------------------------------------------------------------------
// AST Shared Utilities
//
// This file contains shared utility functions for AST traversal and manipulation.
// It provides common operations for finding, filtering, and processing AST nodes

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

func findFuncDecls(root *dst.File, lambda func(*dst.FuncDecl) bool) []*dst.FuncDecl {
	funcDecls := ListFuncDecls(root)

	// The function with receiver and the function without receiver may have
	// the same name, so they need to be classified into the same name
	found := make([]*dst.FuncDecl, 0)
	for _, funcDecl := range funcDecls {
		if lambda(funcDecl) {
			found = append(found, funcDecl)
		}
	}
	return found
}

func FindFuncDeclWithoutRecv(root *dst.File, funcName string) *dst.FuncDecl {
	decls := findFuncDecls(root, func(funcDecl *dst.FuncDecl) bool {
		return funcDecl.Name.Name == funcName && !HasReceiver(funcDecl)
	})

	if len(decls) == 0 {
		return nil
	}
	return decls[0]
}

// stripGenericTypes extracts the base type name from a receiver expression,
// handling both generic and non-generic types.
// For example:
// - *MyStruct -> *MyStruct
// - MyStruct -> MyStruct
// - *GenStruct[T] -> *GenStruct
// - GenStruct[T] -> GenStruct
func stripGenericTypes(recvTypeExpr dst.Expr) string {
	switch expr := recvTypeExpr.(type) {
	case *dst.StarExpr: // func (*Recv)T or func (*Recv[T])T
		// Check if X is an Ident (non-generic) or IndexExpr/IndexListExpr (generic)
		switch x := expr.X.(type) {
		case *dst.Ident:
			// Non-generic pointer receiver: *MyStruct
			return "*" + x.Name
		case *dst.IndexExpr:
			// Generic pointer receiver with single type param: *GenStruct[T]
			if baseIdent, ok := x.X.(*dst.Ident); ok {
				return "*" + baseIdent.Name
			}
		case *dst.IndexListExpr:
			// Generic pointer receiver with multiple type params: *GenStruct[T, U]
			if baseIdent, ok := x.X.(*dst.Ident); ok {
				return "*" + baseIdent.Name
			}
		}
	case *dst.Ident: // func (Recv)T
		return expr.Name
	case *dst.IndexExpr:
		// Generic value receiver with single type param: GenStruct[T]
		if baseIdent, ok := expr.X.(*dst.Ident); ok {
			return baseIdent.Name
		}
	case *dst.IndexListExpr:
		// Generic value receiver with multiple type params: GenStruct[T, U]
		if baseIdent, ok := expr.X.(*dst.Ident); ok {
			return baseIdent.Name
		}
	}
	return ""
}

func FindFuncDecl(root *dst.File, function string, receiverType string) []*dst.FuncDecl {
	decls := findFuncDecls(root, func(funcDecl *dst.FuncDecl) bool {
		return function == funcDecl.Name.Name
	})
	if receiverType != "" {
		filtered := make([]*dst.FuncDecl, 0)
		for _, funcDecl := range decls {
			if !HasReceiver(funcDecl) {
				continue
			}
			re := regexp.MustCompile("^" + receiverType + "$") // strict match
			if !HasReceiver(funcDecl) {
				if re.MatchString("") {
					filtered = append(filtered, funcDecl)
				}
			}

			// Receiver type is specified, and target function has receiver
			// Match both func name and receiver type
			recvTypeExpr := funcDecl.Recv.List[0].Type
			baseType := stripGenericTypes(recvTypeExpr)

			if baseType == "" {
				msg := fmt.Sprintf("unexpected receiver type: %T", recvTypeExpr)
				util.Unimplemented(msg)
			}
			if re.MatchString(baseType) {
				filtered = append(filtered, funcDecl)
			}
		}
		return filtered
	}
	// Receiver type is not specified, return all functions without receiver
	filtered := make([]*dst.FuncDecl, 0)
	for _, funcDecl := range decls {
		if !HasReceiver(funcDecl) {
			filtered = append(filtered, funcDecl)
		}
	}
	return filtered

}

func ListFuncDecls(root *dst.File) []*dst.FuncDecl {
	funcDecls := make([]*dst.FuncDecl, 0)
	for _, decl := range root.Decls {
		funcDecl, ok := decl.(*dst.FuncDecl)
		if !ok {
			continue
		}
		funcDecls = append(funcDecls, funcDecl)
	}
	return funcDecls
}

func FindStructDecl(root *dst.File, structName string) *dst.GenDecl {
	for _, decl := range root.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok && genDecl.Tok == token.TYPE {
			if typeSpec, ok1 := genDecl.Specs[0].(*dst.TypeSpec); ok1 {
				if typeSpec.Name.Name == structName {
					return genDecl
				}
			}
		}
	}
	return nil
}

// SplitMultiNameFields splits fields that have multiple names into separate fields.
// For example, a field like "a, b int" becomes two fields: "a int" and "b int".
func SplitMultiNameFields(fieldList *dst.FieldList) *dst.FieldList {
	if fieldList == nil {
		return nil
	}
	result := &dst.FieldList{List: []*dst.Field{}}
	for _, field := range fieldList.List {
		// Handle unnamed fields (e.g., embedded types) or fields with single/multiple names
		namesToProcess := field.Names
		if len(namesToProcess) == 0 {
			// For unnamed fields, create one field with no names
			namesToProcess = []*dst.Ident{nil}
		}

		for _, name := range namesToProcess {
			clonedType := util.AssertType[dst.Expr](dst.Clone(field.Type))

			var names []*dst.Ident
			if name != nil {
				clonedName := util.AssertType[*dst.Ident](dst.Clone(name))
				names = []*dst.Ident{clonedName}
			}

			newField := &dst.Field{
				Names: names,
				Type:  clonedType,
			}
			result.List = append(result.List, newField)
		}
	}
	return result
}

// CloneTypeParams safely clones a type parameter field list for generic functions.
// Returns nil if the input is nil.
func CloneTypeParams(typeParams *dst.FieldList) *dst.FieldList {
	if typeParams == nil {
		return nil
	}
	return util.AssertType[*dst.FieldList](dst.Clone(typeParams))
}
