package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func sourceMapping(repo string, manifest Manifest) (map[string]string, error) {
	path, err := cleanRepoPath(repo, manifest.Source)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	ast.Inspect(file, func(node ast.Node) bool {
		cc, ok := node.(*ast.CaseClause)
		if !ok {
			return true
		}
		kind, eventType := caseKind(cc), returnEventType(cc)
		if kind != "" && eventType != "" {
			out[kind] = eventType
		}
		return true
	})
	return out, nil
}

func caseKind(cc *ast.CaseClause) string {
	if len(cc.List) != 1 {
		return ""
	}
	sel, ok := cc.List[0].(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return eventKindByConst()[sel.Sel.Name]
}
