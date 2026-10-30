package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func claudeHistoryFixture(t *testing.T) (agentbridge.StartRequest, string) {
	t.Helper()
	req := agentbridge.StartRequest{ResumeSessionID: "12345678-1234-1234-1234-123456789abc", Cwd: "/workspace/repo", Prompt: "# Riido AI Agent Assignment\n- task_id: task\nnext", Metadata: map[string]string{"task_id": "task"}, Env: map[string]string{"CLAUDE_CONFIG_DIR": t.TempDir()}}
	var rows []string
	parent := ""
	for i, content := range []string{req.Prompt, "Riido answer", "external question", "external answer"} {
		id := string(rune('a' + i))
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		row := map[string]any{"uuid": id, "parentUuid": parent, "sessionId": req.ResumeSessionID, "cwd": req.Cwd, "type": role, "timestamp": "2026-10-26T01:00:00Z", "message": map[string]any{"content": []any{map[string]any{"type": "text", "text": content}, map[string]any{"type": "thinking", "thinking": "never expose"}}}}
		data, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, string(data))
		parent = id
	}
	// A later subagent branch must not replace the main conversation leaf.
	rows = append(rows, `{"uuid":"side","parentUuid":"a","isSidechain":true,"type":"assistant","sessionId":"foreign","message":{"content":"private subagent"}}`)
	return req, strings.Join(rows, "\n") + "\n"
}

func TestClaudeReadsExactBoundTranscriptAndPublishesAfterIdentity(t *testing.T) {
	req, data := claudeHistoryFixture(t)
	dir := filepath.Join(req.Env["CLAUDE_CONFIG_DIR"], "projects", historyPathCharacters.ReplaceAllString(req.Cwd, "-"))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, req.ResumeSessionID+".jsonl"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	var imported []agentbridge.Event
	for _, id := range []string{"foreign", req.ResumeSessionID, req.ResumeSessionID} {
		events, _, err := d.OnRaw(t.Context(), agentbridge.RawEvent{Type: "system", Payload: map[string]any{"subtype": "init", "session_id": id}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.NativeHistory != nil {
				imported = append(imported, event)
				if id != req.ResumeSessionID {
					t.Fatal("published under foreign identity")
				}
			}
		}
	}
	if len(imported) != 2 || imported[0].NativeHistory.MessageID != "c" || imported[1].NativeHistory.Body != "external answer" {
		t.Fatalf("import=%+v", imported)
	}
}

func TestClaudeRejectsForeignBrokenAndIncompleteTranscripts(t *testing.T) {
	req, data := claudeHistoryFixture(t)
	for name, bad := range map[string]string{
		"foreign session": strings.ReplaceAll(data, req.ResumeSessionID, "foreign"),
		"foreign project": strings.ReplaceAll(data, req.Cwd, "/other"),
		"broken chain":    strings.Replace(data, `"parentUuid":"c"`, `"parentUuid":"missing"`, 1),
		"partial json":    data + `{"uuid":`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := claudeHistoryMessages(strings.NewReader(bad), req); err == nil {
				t.Fatal("unsafe transcript accepted")
			}
		})
	}
	req.ResumeSessionID = "../../other"
	if _, err := claudeNativeHistory(req); err == nil {
		t.Fatal("unsafe path accepted")
	}
}

func TestClaudeHistoryRetainsImageOnlyUserTurn(t *testing.T) {
	req, data := claudeHistoryFixture(t)
	data = strings.Replace(data, `"type":"text","text":"external question"`, `"type":"image","source":{"data":"private"}`, 1)
	// json.Marshal orders text before type in the fixture's map.
	data = strings.Replace(data, `"text":"external question","type":"text"`, `"type":"image","source":{"data":"private"}`, 1)
	messages, err := claudeHistoryMessages(strings.NewReader(data), req)
	if err != nil || len(messages) != 4 || messages[2].Role != "user" || !strings.Contains(messages[2].Body, "attachment not imported") || strings.Contains(messages[2].Body, "private") {
		t.Fatalf("image-only turn lost or attachment exposed: %+v %v", messages, err)
	}
}
