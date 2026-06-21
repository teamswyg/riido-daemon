package main

import "testing"

func TestEvidenceToolInventoryRejectsUncalledTool(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "tools/lonely/main.go", `package main

func main() {
	_ = "evidence-out"
}
`)
	mustWrite(t, root, ".github/workflows/build.yml", `name: build
jobs:
  build:
    steps:
      - run: go test ./...
`)
	m := manifest{WorkflowRoot: ".github/workflows"}
	got, err := auditWorkflows(root, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MissingEvidenceTools) != 1 || got.MissingEvidenceTools[0] != "lonely" {
		t.Fatalf("missing evidence tools = %#v", got.MissingEvidenceTools)
	}
}

func TestEvidenceToolInventoryRequiresEvidenceOutBinding(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "tools/report/main.go", `package main

func main() {
	_ = "evidence-out"
}
`)
	mustWrite(t, root, ".github/workflows/report.yml", `name: report
jobs:
  report:
    steps:
      - run: go run ./tools/report -check-doc
      - uses: actions/upload-artifact@v4
        with:
          name: report
          path: out/report.json
          if-no-files-found: error
`)
	got, err := auditWorkflows(root, manifest{WorkflowRoot: ".github/workflows"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MissingEvidenceTools) != 0 {
		t.Fatalf("missing evidence tools = %#v", got.MissingEvidenceTools)
	}
	if len(got.MissingEvidenceToolBindings) != 1 ||
		got.MissingEvidenceToolBindings[0] != "report" {
		t.Fatalf("missing evidence bindings = %#v", got.MissingEvidenceToolBindings)
	}
}

func mustWrite(t *testing.T, root, path, text string) {
	t.Helper()
	if err := writeText(repoPath(root, path), text); err != nil {
		t.Fatal(err)
	}
}
