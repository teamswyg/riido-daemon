package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func draftFields(repo string, manifest Manifest) (map[string]struct{}, error) {
	path, err := cleanRepoPath(repo, manifest.DraftSource)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	fields := map[string]struct{}{}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Draft" {
			return true
		}
		st, ok := spec.Type.(*ast.StructType)
		if !ok {
			return false
		}
		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				fields[name.Name] = struct{}{}
			}
		}
		return false
	})
	return fields, nil
}
