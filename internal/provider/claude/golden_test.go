package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// loadFixtureLines reads a JSONL file and returns one RawEvent per
// non-empty line. The file path is resolved relative to the package's
// testdata directory.
func loadFixtureLines(t *testing.T, name string) []agentbridge.RawEvent {
	t.Helper()
	path := filepath.Join("testdata", name)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var out []agentbridge.RawEvent
	scanner := bufio.NewScanner(f)
	// Allow up to 10 MB per line — matches MaxLineBytes in the parser.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("%s: parse %q: %v", path, line, err)
		}
		typ, _ := m["type"].(string)
		out = append(out, agentbridge.RawEvent{
			Source:  agentbridge.RawSourceStdout,
			Type:    typ,
			Payload: m,
			Bytes:   []byte(line),
		})
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return out
}

// TestGoldenNormalJSONL processes every line of testdata/normal.jsonl
// through Translate and asserts that the full sequence yields at least
// one SessionIdentified, one TextDelta, and one Result.
func TestGoldenNormalJSONL(t *testing.T) {
	raws := loadFixtureLines(t, "normal.jsonl")
	if len(raws) == 0 {
		t.Fatal("fixture empty")
	}

	var saw struct {
		session bool
		text    bool
		result  bool
	}
	for _, raw := range raws {
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatalf("translate %+v: %v", raw, err)
		}
		for _, ev := range evs {
			switch ev.Kind {
			case agentbridge.EventSessionIdentified:
				saw.session = true
			case agentbridge.EventTextDelta:
				saw.text = true
			case agentbridge.EventResult:
				saw.result = true
			}
		}
	}
	if !saw.session || !saw.text || !saw.result {
		t.Fatalf("fixture coverage gap: %+v", saw)
	}
}

// TestGoldenToolUseJSONL processes tool_use.jsonl and asserts on the
// tool lifecycle events.
func TestGoldenToolUseJSONL(t *testing.T) {
	raws := loadFixtureLines(t, "tool_use.jsonl")
	var started, completed, failed bool
	for _, raw := range raws {
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range evs {
			switch ev.Kind {
			case agentbridge.EventToolCallStarted:
				started = true
			case agentbridge.EventToolCallCompleted:
				completed = true
			case agentbridge.EventToolCallFailed:
				failed = true
			}
		}
	}
	if !started || !completed || !failed {
		t.Fatalf("tool_use fixture coverage gap: started=%v completed=%v failed=%v", started, completed, failed)
	}
}

// TestGoldenControlRequestJSONL processes control_request.jsonl and
// asserts that approval needed + error result are surfaced.
func TestGoldenControlRequestJSONL(t *testing.T) {
	raws := loadFixtureLines(t, "control_request.jsonl")
	var approval bool
	var failedResult bool
	for _, raw := range raws {
		evs, _, _ := Translate(raw)
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventToolApprovalNeeded {
				approval = true
			}
			if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultFailed {
				failedResult = true
			}
		}
	}
	if !approval || !failedResult {
		t.Fatalf("control_request fixture coverage gap: approval=%v failed=%v", approval, failedResult)
	}
}
