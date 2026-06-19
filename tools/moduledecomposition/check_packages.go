package main

func checkPackageRoles(m manifest, packages map[string]packageInfo) []checkResult {
	var results []checkResult
	for _, role := range m.PackageRoles {
		for _, rel := range role.Packages {
			name := "package-exists:" + rel
			importPath := relativeImport(m, rel)
			results = append(results, checkResult{Name: name, File: rel, Pass: hasPackage(packages, importPath)})
		}
	}
	return results
}

func checkBinaryPackage(m manifest, packages map[string]packageInfo) []checkResult {
	importPath := relativeImport(m, m.BinaryPackage)
	pkg, ok := packages[importPath]
	pass := ok && pkg.Name == "main"
	detail := ""
	if ok && pkg.Name != "main" {
		detail = "package name is " + pkg.Name
	}
	return []checkResult{{Name: "binary-package", File: m.BinaryPackage, Pass: pass, Detail: detail}}
}
