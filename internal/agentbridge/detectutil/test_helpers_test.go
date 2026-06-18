package detectutil

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func containsDir(dirs []string, want string) bool {
	return slices.Contains(dirs, want)
}

func envListValue(env []string, wantKey string) (string, bool) {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, wantKey) {
			return value, true
		}
	}
	return "", false
}

func writeExecutable(t *testing.T, path, output string) string {
	t.Helper()
	script := "#!/bin/sh\necho '" + output + "'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
	return filepath.Clean(path)
}

func overrideAugmentedSearchDirs(t *testing.T, dirs ...string) {
	t.Helper()
	restore := augmentedSearchDirs
	augmentedSearchDirs = func() []string { return dirs }
	t.Cleanup(func() { augmentedSearchDirs = restore })
}
