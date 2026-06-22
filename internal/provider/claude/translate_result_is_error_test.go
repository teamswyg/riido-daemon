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
	if !strings.Contains(last.Result.Error, "Not logged in") {
		t.Fatalf("error: %q", last.Result.Error)
	}
}
