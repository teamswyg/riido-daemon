package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func resultStatusValues(repo string, manifest Manifest) (map[string]string, error) {
	path, err := cleanRepoPath(repo, manifest.Sources.ResultStatus)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || statusTypeName(spec.Type) != "ResultStatus" {
			return true
		}
		for i, name := range spec.Names {
			if i < len(spec.Values) {
				out[name.Name] = stringLiteral(spec.Values[i])
			}
		}
		return true
	})
	return out, nil
}

func statusTypeName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func stringLiteral(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		return ""
	}
	value, _ := strconv.Unquote(lit.Value)
	return value
}
