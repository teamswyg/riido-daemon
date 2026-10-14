package main

import "fmt"

func run(opts options) error {
	opts = normalizeOptions(opts)
	m, err := loadManifest(opts.Repo, opts.Manifest)
	if err != nil {
		return err
	}
	packages, goListChecks := loadPackages(opts.Repo)
	packageChecks := goListChecks
	packageChecks = append(packageChecks, checkBinaryPackage(m, packages)...)
	packageChecks = append(packageChecks, checkPackageRoles(m, packages)...)
	importChecks := checkImportRules(m, packages)
	problems := validateManifest(m)
	problems = append(problems, failedChecks("package check failed", packageChecks)...)
	problems = append(problems, failedChecks("import check failed", importChecks)...)
	docs := renderedIfValid(m, problems)
	if err := maybeWriteDocs(opts, docs); err != nil {
		problems = append(problems, problem{Message: err.Error()})
	}
	problems = append(problems, checkDocs(opts, docs)...)
	if opts.EvidenceOut != "" {
		if err := writeJSON(opts.EvidenceOut, buildEvidence(m, problems, packageChecks, importChecks)); err != nil {
			return err
		}
	}
	if len(problems) > 0 {
		return problemError(problems)
	}
	fmt.Println("module-decomposition-docs: clean")
	return nil
}

func renderedIfValid(m manifest, problems []problem) map[string]string {
	if len(problems) > 0 {
		return map[string]string{}
	}
	return renderedDocs(m)
}
