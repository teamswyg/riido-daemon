package codex

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func loadGoldenFixtureLines(t *testing.T, name string) []agentbridge.RawEvent {
	t.Helper()
	path := filepath.Join("testdata", name)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	var out []agentbridge.RawEvent
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		out = append(out, parseCodexGoldenLine(t, path, line))
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return out
}

func parseCodexGoldenLine(t *testing.T, path, line string) agentbridge.RawEvent {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("%s: parse %q: %v", path, line, err)
	}
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    classifyJSONRPC(payload),
		Payload: payload,
		Bytes:   []byte(line),
	}
}
