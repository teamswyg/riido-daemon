package main

import "go/ast"

func emptyStatusDefault(file *ast.File) string {
	found := ""
	ast.Inspect(file, func(node ast.Node) bool {
		ifs, ok := node.(*ast.IfStmt)
		if !ok || !isEmptyStatusCond(ifs.Cond) {
			return true
		}
		for _, stmt := range ifs.Body.List {
			if name := assignedSelector(stmt); name != "" {
				found = name
			}
		}
		return true
	})
	return found
}

func isEmptyStatusCond(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	left, lok := bin.X.(*ast.Ident)
	right, rok := bin.Y.(*ast.BasicLit)
	return lok && rok && left.Name == "status" && right.Value == `""`
}

func assignedSelector(stmt ast.Stmt) string {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok || len(assign.Rhs) != 1 {
		return ""
	}
	return selectorName(assign.Rhs[0])
}
