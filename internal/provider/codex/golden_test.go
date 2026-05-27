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

// loadFixtureLines reads a JSONL of JSON-RPC frames and returns them as
// classified RawEvents (response / notification:<method> /
// server_request:<method>).
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
		out = append(out, agentbridge.RawEvent{
			Source:  agentbridge.RawSourceStdout,
			Type:    classifyJSONRPC(m),
			Payload: m,
			Bytes:   []byte(line),
		})
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return out
}

// TestGoldenNormalRPC asserts the full normal turn fixture produces the
// expected lifecycle: SessionIdentified → Lifecycle → text/reasoning →
// tool lifecycle → usage → result(Completed).
func TestGoldenNormalRPC(t *testing.T) {
	raws := loadFixtureLines(t, "normal_rpc.jsonl")
	var saw struct {
		session, lifecycle, text, reasoning, toolStart, toolDone, usage, result bool
	}
	for _, raw := range raws {
		evs, _, err := Translate(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range evs {
			switch ev.Kind {
			case agentbridge.EventSessionIdentified:
				saw.session = true
			case agentbridge.EventLifecycle:
				saw.lifecycle = true
			case agentbridge.EventTextDelta:
				saw.text = true
			case agentbridge.EventThinkingDelta:
				saw.reasoning = true
			case agentbridge.EventToolCallStarted:
				saw.toolStart = true
			case agentbridge.EventToolCallCompleted:
				saw.toolDone = true
			case agentbridge.EventUsageDelta:
				saw.usage = true
			case agentbridge.EventResult:
				if ev.Result.Status == agentbridge.ResultCompleted {
					saw.result = true
				}
			}
		}
	}
	if !(saw.session && saw.lifecycle && saw.text && saw.reasoning && saw.toolStart && saw.toolDone && saw.usage && saw.result) {
		t.Fatalf("normal_rpc coverage gap: %+v", saw)
	}
}

// TestGoldenApprovalRPC asserts the approval fixture produces
// ToolApprovalNeeded events for both approve_command and approve_patch
// server_requests, and a failed result.
func TestGoldenApprovalRPC(t *testing.T) {
	raws := loadFixtureLines(t, "approval_rpc.jsonl")
	approvals := 0
	failedResult := false
	for _, raw := range raws {
		evs, _, _ := Translate(raw)
		for _, ev := range evs {
			if ev.Kind == agentbridge.EventToolApprovalNeeded {
				approvals++
			}
			if ev.Kind == agentbridge.EventResult && ev.Result.Status == agentbridge.ResultFailed {
				failedResult = true
			}
		}
	}
	if approvals != 2 {
		t.Fatalf("expected 2 approval events (command + patch), got %d", approvals)
	}
	if !failedResult {
		t.Fatalf("approval_rpc fixture missing failed result event")
	}
}
