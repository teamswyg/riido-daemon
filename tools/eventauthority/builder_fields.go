package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func builderFields(repo string, manifest Manifest) (map[string]struct{}, error) {
	path, err := cleanRepoPath(repo, manifest.BuilderSource)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	fields := map[string]struct{}{}
	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok || !isCanonicalEventLit(lit) {
			return true
		}
		for _, elt := range lit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok {
					fields[ident.Name] = struct{}{}
				}
			}
		}
		return false
	})
	return fields, nil
}

func isCanonicalEventLit(lit *ast.CompositeLit) bool {
	sel, ok := lit.Type.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "CanonicalEvent"
}
