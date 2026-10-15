package cursor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCursorErrorResultNeverCompletes(t *testing.T) {
	for _, body := range []string{
		`{"type":"result","subtype":"success","is_error":true,"result":"Login required"}`,
		`{"type":"result","subtype":"error_max_turns","result":"Turn limit reached"}`,
		`{"type":"result","subtype":"success","errors":["Quota exhausted"]}`,
	} {
		events := tx(t, rawJSON(t, body))
		res := events[len(events)-1].Result
		if res.Status != agentbridge.ResultFailed || res.Error == "" {
			t.Fatalf("native failure accepted: %+v", res)
		}
	}
}

func TestCursorSystemInstructionKeepsTaskFirst(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{Prompt: "Fix calculator", SystemPrompt: "Return the test result"}, StartOptions{})
	if err != nil || len(cmd.Args) < 2 || !strings.HasPrefix(cmd.Args[1], "Fix calculator") || !strings.Contains(cmd.Args[1], "Return the test result") {
		t.Fatalf("task/instruction delivery: %+v, %v", cmd, err)
	}
}
