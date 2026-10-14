package supervisor

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func readNativeConfigManifest(t *testing.T, path string) workdir.NativeConfigManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read native config manifest: %v", err)
	}
	var manifest workdir.NativeConfigManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode native config manifest: %v", err)
	}
	return manifest
}

func containsString(values []string, want string) bool {
	return slices.Contains(values, want)
}
