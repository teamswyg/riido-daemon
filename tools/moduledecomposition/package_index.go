package main

import "strings"

func relativeImport(m manifest, rel string) string {
	return m.ModulePath + "/" + rel
}

func hasPackage(packages map[string]packageInfo, importPath string) bool {
	_, ok := packages[importPath]
	return ok
}

func matchingPackages(m manifest, packages map[string]packageInfo, prefixes []string) []packageInfo {
	var out []packageInfo
	for _, pkg := range packages {
		rel := strings.TrimPrefix(pkg.ImportPath, m.ModulePath+"/")
		for _, prefix := range prefixes {
			if rel == prefix || strings.HasSuffix(prefix, "/") && strings.HasPrefix(rel, prefix) {
				out = append(out, pkg)
				break
			}
		}
	}
	return out
}
