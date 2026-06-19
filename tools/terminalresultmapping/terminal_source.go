package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type terminalSource struct {
	Cases        map[string]string
	Fallback     string
	EmptyDefault string
}

func terminalSourceMapping(repo string, manifest Manifest) (terminalSource, error) {
	path, err := cleanRepoPath(repo, manifest.Sources.TerminalResult)
	if err != nil {
		return terminalSource{}, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return terminalSource{}, err
	}
	out := terminalSource{Cases: map[string]string{}, EmptyDefault: emptyStatusDefault(file)}
	ast.Inspect(file, func(node ast.Node) bool {
		cc, ok := node.(*ast.CaseClause)
		if !ok {
			return true
		}
		eventType := returnEventType(cc)
		if len(cc.List) == 0 && eventType != "" {
			out.Fallback = eventType
			return true
		}
		for _, expr := range cc.List {
			if name := selectorName(expr); name != "" && eventType != "" {
				out.Cases[name] = eventType
			}
		}
		return true
	})
	return out, nil
}

func selectorName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return sel.Sel.Name
}
