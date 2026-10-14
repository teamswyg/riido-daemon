package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
)

func interfaceMethods(repo string, spec InterfaceSpec) ([]string, error) {
	path, err := cleanRepoPath(repo, spec.File)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	methods := []string{}
	ast.Inspect(file, func(node ast.Node) bool {
		ts, ok := node.(*ast.TypeSpec)
		if !ok || ts.Name.Name != spec.Name {
			return true
		}
		if iface, ok := ts.Type.(*ast.InterfaceType); ok {
			methods = append(methods, methodNames(iface)...)
		}
		return true
	})
	slices.Sort(methods)
	return methods, nil
}

func methodNames(iface *ast.InterfaceType) []string {
	out := []string{}
	for _, field := range iface.Methods.List {
		for _, name := range field.Names {
			out = append(out, name.Name)
		}
	}
	return out
}
