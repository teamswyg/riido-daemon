package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryValidatesClaimBindings(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, ".pre-commit-config.yaml", "loop-registry\n"+defaultCommand())
	writeFixture(t, root, ".github/workflows/loop-registry.yml", defaultCommand())
	writeFixture(t, root, "code.go", "package fixture\n")
	writeFixture(t, root, "doc.md", "doc")
	writeFixture(t, root, "code_test.go", "TestClaimBinding")
	writeFixture(t, root, defaultManifest, fixtureManifest())
	chdir(t, root)

	if err := run(options{Manifest: defaultManifest, CheckDoc: false}); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryRejectsRuntimeOnlyClaimChange(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module fixture\n")
	writeFixture(t, root, ".pre-commit-config.yaml", "loop-registry\n"+defaultCommand())
	writeFixture(t, root, ".github/workflows/loop-registry.yml", defaultCommand())
	writeFixture(t, root, "code.go", "package fixture\n")
	writeFixture(t, root, "doc.md", "doc")
	writeFixture(t, root, "code_test.go", "TestClaimBinding")
	writeFixture(t, root, "changed.txt", "code.go\n")
	writeFixture(t, root, defaultManifest, fixtureManifest())
	chdir(t, root)

	err := run(options{Manifest: defaultManifest, ChangedFiles: "changed.txt"})
	if err == nil {
		t.Fatal("expected runtime-only claim change to fail")
	}
}

func writeFixture(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}
