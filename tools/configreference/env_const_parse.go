package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func parseEnvConstants(path string) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		collectConstSpecStrings(out, gen.Specs)
	}
	return out, nil
}

func collectConstSpecStrings(out map[string]string, specs []ast.Spec) {
	for _, spec := range specs {
		value, ok := constStringValue(spec)
		if !ok {
			continue
		}
		for _, name := range spec.(*ast.ValueSpec).Names {
			if strings.HasPrefix(name.Name, "env") {
				out[name.Name] = value
			}
		}
	}
}

func constStringValue(spec ast.Spec) (string, bool) {
	valueSpec, ok := spec.(*ast.ValueSpec)
	if !ok || len(valueSpec.Values) != 1 {
		return "", false
	}
	lit, ok := valueSpec.Values[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	return value, err == nil
}
