package surfacecontract

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const maxVisibleFileLines = 200

var generatedPartName = regexp.MustCompile(`(?i)(^|[_-])part[_-]?[0-9]+`)

func TestRepositorySurfaceAreaContract(t *testing.T) {
	root := repoRoot(t)
	var partNamedFiles []string
	var oversizedFiles []string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if ignoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		rel := mustRel(t, root, path)
		if generatedPartName.MatchString(filepath.Base(path)) {
			partNamedFiles = append(partNamedFiles, rel)
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if isBinary(body) {
			return nil
		}
		if lines(body) > maxVisibleFileLines {
			oversizedFiles = append(oversizedFiles, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(partNamedFiles) > 0 {
		t.Fatalf("semantic filenames are required; replace part-numbered files:\n%s", strings.Join(partNamedFiles, "\n"))
	}
	if len(oversizedFiles) > 0 {
		t.Fatalf("visible files must stay under %d lines:\n%s", maxVisibleFileLines, strings.Join(oversizedFiles, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func mustRel(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

func ignoredDirectory(name string) bool {
	switch name {
	case ".git", ".cache", ".mypy_cache", ".pytest_cache", ".ruff_cache", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func isBinary(body []byte) bool {
	sample := body
	if len(sample) > 8192 {
		sample = sample[:8192]
	}
	return bytes.IndexByte(sample, 0) >= 0
}

func lines(body []byte) int {
	if len(body) == 0 {
		return 0
	}
	count := bytes.Count(body, []byte{'\n'})
	if body[len(body)-1] != '\n' {
		count++
	}
	return count
}
