package main

import "go/ast"

func returnEventType(cc *ast.CaseClause) string {
	for _, stmt := range cc.Body {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok || len(ret.Results) == 0 {
			continue
		}
		if sel, ok := ret.Results[0].(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
	}
	return ""
}
