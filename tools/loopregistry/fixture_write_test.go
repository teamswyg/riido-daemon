package main

import "testing"

func writeLoopRegistryFixture(t *testing.T, root string) {
	t.Helper()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, ".pre-commit-config.yaml", "loop-registry\n"+defaultCommand())
	writeFixture(t, root, ".github/workflows/loop-registry.yml", fixtureWorkflow())
	writeFixture(t, root, "code.go", "package fixture\n")
	writeFixture(t, root, "doc.md", "doc")
	writeFixture(t, root, "code_test.go", "TestClaimBinding")
	writeFixture(t, root, defaultManifest, fixtureManifest())
}
