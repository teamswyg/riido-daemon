package openclaw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeOpenClawConfigFixture(t *testing.T, workspace, model string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openclaw.json")
	if err := os.WriteFile(path, []byte(openClawConfigFixture(workspace, model)), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func openClawConfigFixture(workspace, model string) string {
	return `{"agents":{"defaults":{"workspace":"` + workspace + `","model":{"primary":"` +
		model + `"}},"list":[{"id":"main","model":"` + model + `"}]}}`
}

func envValueFromList(env []string, key string) string {
	for _, entry := range env {
		if value, ok := strings.CutPrefix(entry, key+"="); ok {
			return value
		}
	}
	return ""
}
