package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitCommaTrimsAndDropsEmptyValues(t *testing.T) {
	got := splitComma(" docs/a.md, ,tools/a_test.go,\ninternal/a.go ")
	want := []string{"docs/a.md", "tools/a_test.go", "internal/a.go"}
	if len(got) != len(want) {
		t.Fatalf("split count=%v got=%v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("split[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestGitOutputWithWorkTreeReadsDetachedWorkTree(t *testing.T) {
	repo := initGitFixture(t)
	writeGitFile(t, repo, "tracked.txt", "changed")
	out, ok := gitOutputWithWorkTree(repo, "diff", "--name-only", "HEAD")
	if !ok || !strings.Contains(out, "tracked.txt") {
		t.Fatalf("git output ok=%v out=%q", ok, out)
	}
}

func TestValidatePathsReportsEmptyMissingAndAcceptsExisting(t *testing.T) {
	repo := t.TempDir()
	writeFixtureFile(t, repo, "docs/claim.md")
	problems := validatePaths(repo, "claim-a", []string{
		"",
		"docs/claim.md",
		"docs/missing.md",
	})
	if len(problems) != 2 {
		t.Fatalf("problems=%#v", problems)
	}
	got := problems[0].Message + "\n" + problems[1].Message
	for _, want := range []string{"contains empty path", "missing path docs/missing.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestLoadManifestReadsAbsolutePath(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "manifest.json")
	data := []byte(`{"schema_version":"riido-semantic-change-bindings.v1","id":"m"}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadManifest("ignored", path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "m" {
		t.Fatalf("manifest=%+v", got)
	}
}
