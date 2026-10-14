package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func sourceEventKinds(repo string) (map[string]struct{}, error) {
	path, err := cleanRepoPath(repo, "internal/agentbridge/event_kind.go")
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	kinds := map[string]struct{}{}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || !isEventKindSpec(spec) {
			return true
		}
		for _, value := range spec.Values {
			if lit, ok := value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if decoded, err := strconv.Unquote(lit.Value); err == nil {
					kinds[decoded] = struct{}{}
				}
			}
		}
		return true
	})
	if len(kinds) == 0 {
		return nil, fmt.Errorf("no EventKind source declarations found")
	}
	return kinds, nil
}

func isEventKindSpec(spec *ast.ValueSpec) bool {
	ident, ok := spec.Type.(*ast.Ident)
	return ok && ident.Name == "EventKind"
}
