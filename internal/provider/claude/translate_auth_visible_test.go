package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const claude401Text = "Failed to authenticate, API Error: 401 Invalid authentication credentials"

func TestTranslateLogClaude401GuidesReauthentication(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"log","message":"`+claude401Text+`"}`)
	events := translate(t, raw)

	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("events: %+v", events)
	}
	assertClaudeAuthGuide(t, events[0].Text)
}

func TestTranslateStderrClaude401GuidesReauthentication(t *testing.T) {
	raw := agentbridge.RawEvent{
		Source: agentbridge.RawSourceStderr,
		Type:   "stderr",
		Bytes:  []byte(claude401Text),
	}
	events := translate(t, raw)

	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("events: %+v", events)
	}
	assertClaudeAuthGuide(t, events[0].Text)
}

func TestTranslateAssistantTextClaude401GuidesReauthentication(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"assistant","message":{"content":[{"type":"text","text":"`+claude401Text+`"}]}}`)
	events := translate(t, raw)

	if len(events) != 1 || events[0].Kind != agentbridge.EventTextDelta {
		t.Fatalf("events: %+v", events)
	}
	assertClaudeAuthGuide(t, events[0].Text)
}

func assertClaudeAuthGuide(t *testing.T, got string) {
	t.Helper()
	if strings.Contains(got, "Invalid authentication credentials") {
		t.Fatalf("raw auth error leaked: %q", got)
	}
	if !strings.Contains(got, "Claude Code 인증") ||
		!strings.Contains(got, "claude auth login") {
		t.Fatalf("missing auth recovery guide: %q", got)
	}
}
