package main

import "testing"

func TestRegistryRejectsUncoveredClaimWorkflowPath(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, ".pre-commit-config.yaml", "loop-registry\n"+defaultCommand())
	writeFixture(t, root, ".github/workflows/loop-registry.yml", defaultCommand())
	writeFixture(t, root, "code.go", "package fixture\n")
	writeFixture(t, root, "doc.md", "doc")
	writeFixture(t, root, "code_test.go", "TestClaimBinding")
	writeFixture(t, root, defaultManifest, fixtureManifest())
	chdir(t, root)

	err := run(options{Manifest: defaultManifest})
	if err == nil {
		t.Fatal("expected uncovered claim path to fail")
	}
}

func fixtureWorkflow() string {
	return `on:
  pull_request:
    paths:
      - "code.go"
      - "doc.md"
      - "code_test.go"
jobs:
  loop-registry:
    steps:
      - run: ` + defaultCommand()
}
