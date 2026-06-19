package main

import "strings"

func checkImportRules(m manifest, packages map[string]packageInfo) []checkResult {
	var results []checkResult
	for _, rule := range m.ImportRules {
		results = append(results, checkImportRule(m, packages, rule))
	}
	return results
}

func checkImportRule(m manifest, packages map[string]packageInfo, rule importRule) checkResult {
	result := checkResult{Name: "import-rule:" + rule.Group, Pass: true}
	for _, pkg := range matchingPackages(m, packages, rule.PackagePrefixes) {
		for _, imported := range pkg.Imports {
			for _, forbidden := range rule.ForbiddenPrefixes {
				if strings.HasPrefix(imported, forbidden) {
					result.Pass = false
					result.File = strings.TrimPrefix(pkg.ImportPath, m.ModulePath+"/")
					result.Detail = imported
					return result
				}
			}
		}
	}
	return result
}
