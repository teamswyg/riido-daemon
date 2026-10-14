package codex

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func jsonline(t *testing.T, m map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return append(b, '\n')
}

func mustWriteJSONRPC(t *testing.T, r *process.FakeRunning, m map[string]any) {
	t.Helper()
	if err := r.WriteStdin(jsonline(t, m)); err != nil {
		t.Fatalf("WriteStdin: %v", err)
	}
}
