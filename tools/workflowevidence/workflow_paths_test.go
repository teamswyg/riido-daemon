package main

import "testing"

func TestWorkflowInventoryIncludesYamlWorkflows(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "tools/report/main.go", `package main

func main() {
	_ = "evidence-out"
}
`)
	mustWrite(t, root, ".github/workflows/report.yaml", `name: report
jobs:
  report:
    steps:
      - run: go run ./tools/report -evidence-out out/report.json
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
	if len(got.Records) != 1 || got.Records[0].Path != ".github/workflows/report.yaml" {
		t.Fatalf("records = %#v", got.Records)
	}
	if got.Covered != 1 || got.EvidenceToolBound != 1 {
		t.Fatalf("coverage = %+v", got)
	}
}
