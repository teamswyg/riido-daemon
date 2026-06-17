package agentexecutionevidence

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertTestExists(t *testing.T, dir, testName string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read package dir %s: %v", dir, err)
	}
	needle := "func " + testName + "("
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if strings.Contains(readText(t, filepath.Join(dir, entry.Name())), needle) {
			return
		}
	}
	t.Fatalf("test %s not found under %s", testName, dir)
}

func assertDocMentionsTest(t *testing.T, docText, testName string) {
	t.Helper()
	if !strings.Contains(docText, testName) {
		t.Fatalf("human doc must mention evidence test %q", testName)
	}
}
