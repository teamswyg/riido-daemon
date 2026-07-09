package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestScanArtifactRootsReportsShapeAndContentProblems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "not-dir"), "file")
	writeFile(t, filepath.Join(root, "store", "nested", "claude"), "binary")
	writeFile(t, filepath.Join(root, "store", "config.txt"), "/Users/teddy/.riido\n")

	problems := strings.Join(scanArtifactRoots(root, []string{
		"missing",
		"not-dir",
		"store",
	}, []string{"claude"}), "\n")

	for _, wanted := range []string{
		"store artifact root missing: missing",
		"store artifact root is not a directory: not-dir",
		"provider CLI appears bundled in store artifact root:",
		"store artifact contains hardcoded user path:",
	} {
		if !strings.Contains(problems, wanted) {
			t.Fatalf("missing %q in %q", wanted, problems)
		}
	}
}

func TestScanArtifactRootsAcceptsCleanArtifacts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "store", "README.md"), "portable artifact\n")
	if problems := scanArtifactRoots(root, []string{"store"}, []string{"claude"}); len(problems) != 0 {
		t.Fatalf("clean store artifacts should pass: %v", problems)
	}
}
