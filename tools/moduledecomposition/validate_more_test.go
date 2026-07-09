package main

import (
	"strings"
	"testing"
)

func TestPackageAndImportChecksProduceEvidence(t *testing.T) {
	m := manifest{ModulePath: "example.com/repo", BinaryPackage: "cmd/app",
		PackageRoles: []packageRole{{Packages: []string{"internal/core"}}},
		ImportRules: []importRule{{
			Group: "core", PackagePrefixes: []string{"internal/core"},
			ForbiddenPrefixes: []string{"example.com/repo/cmd/"},
		}},
	}
	packages := map[string]packageInfo{
		"example.com/repo/cmd/app": {
			ImportPath: "example.com/repo/cmd/app", Name: "main",
		},
		"example.com/repo/internal/core": {
			ImportPath: "example.com/repo/internal/core", Name: "core",
			Imports: []string{"example.com/repo/cmd/app"},
		},
	}
	if checks := checkBinaryPackage(m, packages); len(checks) != 1 || !checks[0].Pass {
		t.Fatalf("binary checks = %+v", checks)
	}
	if checks := checkPackageRoles(m, packages); len(checks) != 1 || !checks[0].Pass {
		t.Fatalf("role checks = %+v", checks)
	}
	importChecks := checkImportRules(m, packages)
	if len(importChecks) != 1 || importChecks[0].Pass || importChecks[0].Detail == "" {
		t.Fatalf("import checks = %+v", importChecks)
	}
}

func TestBuildEvidenceValidateAndProblemMessages(t *testing.T) {
	m := manifest{ID: "modules", GeneratedDoc: "root.md",
		DetailDocs: []detailDoc{{Path: "a"}, {Path: "b"}, {Path: "c"}, {Path: "d"}, {Path: "e"}},
	}
	problems := []problem{{Message: "bad"}}
	ev := buildEvidence(m, problems, []checkResult{{Name: "pkg"}}, nil)
	if ev.Status != "failed" || len(ev.GeneratedDocs) != 6 || len(ev.PackageChecks) != 1 {
		t.Fatalf("evidence = %+v", ev)
	}
	if !strings.Contains(problemError(problems).Error(), "bad") {
		t.Fatal("problem error should include message")
	}
	if rendered := renderedIfValid(m, problems); len(rendered) != 0 {
		t.Fatalf("rendered = %+v", rendered)
	}
}
