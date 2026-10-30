package history

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestExternalSelectsOnlyBoundExternalTurns(t *testing.T) {
	req := agentbridge.StartRequest{ResumeSessionID: "session", Prompt: assignmentPrefix + "next", Metadata: map[string]string{"task_id": "task"}}
	messages := []agentbridge.NativeHistoryMessage{}
	for i, row := range [][2]string{{"user", "pre-Riido private input"}, {"agent", "private answer"}, {"user", assignmentPrefix + "- task_id: task\n"}, {"agent", "already projected"}, {"user", "external question"}, {"agent", "same answer"}, {"user", "external followup"}, {"agent", "same answer"}, {"user", assignmentPrefix + "- task_id: task\n"}, {"agent", "already projected again"}} {
		messages = append(messages, agentbridge.NativeHistoryMessage{SessionID: "session", MessageID: string(rune('a' + i)), Role: row[0], Body: row[1]})
	}
	messages = append(messages, messages[7])
	events, err := External(req, messages)
	if err != nil || len(events) != 4 {
		t.Fatalf("visible external history: %d, %v", len(events), err)
	}
	for i, event := range events {
		if event.NativeHistory.MessageID != messages[i+4].MessageID || event.NativeHistory.Ordinal != i+5 {
			t.Fatalf("source order/identity lost: %+v", event)
		}
	}
	if !Enabled(req) || Enabled(agentbridge.StartRequest{Prompt: req.Prompt}) {
		t.Fatal("only an explicit Riido resume may read history")
	}
	for _, name := range []string{"session", "task", "anchor", "oversize"} {
		t.Run(name, func(t *testing.T) {
			bad := append([]agentbridge.NativeHistoryMessage(nil), messages...)
			switch name {
			case "session":
				bad[4].SessionID = "foreign"
			case "task":
				bad[2].Body = assignmentPrefix + "- task_id: foreign\n"
			case "anchor":
				bad = bad[4:8]
			case "oversize":
				bad[4].Body = strings.Repeat("x", 65537)
			}
			if got, err := External(req, bad); err == nil || got != nil {
				t.Fatal("invalid or partial history was returned")
			}
		})
	}
}
