package figmaboundary

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return body
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("missing %q", want)
	}
}

func requireSliceContains(t *testing.T, items []string, want string) {
	t.Helper()
	if slices.Contains(items, want) {
		return
	}
	t.Fatalf("missing %q in %#v", want, items)
}

func requireSameStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("string slice length = %d, want %d: got %#v want %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("string slice[%d] = %q, want %q: got %#v want %#v", i, got[i], want[i], got, want)
		}
	}
}

func requireToolLimitation(t *testing.T, limitations []toolLimitation, sourceID string) toolLimitation {
	t.Helper()
	for _, limitation := range limitations {
		if limitation.SourceID == sourceID {
			return limitation
		}
	}
	t.Fatalf("missing mirrored tool limitation %q", sourceID)
	return toolLimitation{}
}
