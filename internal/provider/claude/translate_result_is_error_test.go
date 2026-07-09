package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateResultSuccessSubtypeWithIsErrorFails(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"success","is_error":true,"result":"Not logged in"}`)
	events := translate(t, raw)
	last := events[len(events)-1]

	if last.Kind != agentbridge.EventResult ||
		last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("result: %+v", last)
	}
	if !strings.Contains(last.Result.Error, "claude auth login") {
		t.Fatalf("error: %q", last.Result.Error)
	}
}

func TestTranslateResultClaude401GuidesReauthentication(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"result","subtype":"error","error":"Failed to authenticate, API Error: 401 Invalid authentication credentials"}`)
	events := translate(t, raw)
	last := events[len(events)-1]

	if last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("status: %s", last.Result.Status)
	}
	if !strings.Contains(last.Result.Error, "Claude Code 인증") ||
		!strings.Contains(last.Result.Error, "claude auth status") ||
		!strings.Contains(last.Result.Error, "claude auth login") {
		t.Fatalf("error: %q", last.Result.Error)
	}
}
