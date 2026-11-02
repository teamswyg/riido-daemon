package contextbroker

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func changesStatesResult(changes string, complete bool) Result {
	flags := `"complete":true,"full_refresh_required":false`
	if !complete {
		flags = `"complete":false,"full_refresh_required":true`
	}
	return Result{CapturedAt: time.Now(), Data: json.RawMessage(`{"snapshot_id":"asg-1","since":"previous","changes":` + changes + `,` + flags + `}`)}
}

func TestChangesAcceptExplainedOmissionsWithoutMetadata(t *testing.T) {
	changes := `[{"resource_kind":"component","source_id":"gone","part":"editor","status":"deleted"},{"resource_kind":"meeting","source_id":"moved","part":"board","status":"unselected"}]`
	env := readChangesTest(t, changesStatesResult(changes, true), "")
	raw, _ := json.Marshal(env.Data)
	if env.Error != nil || !strings.Contains(string(raw), `"status":"deleted"`) || !strings.Contains(string(raw), `"status":"unselected"`) || !strings.Contains(string(raw), `"complete":true`) {
		t.Fatalf("explained omissions rejected: %v %s", env.Error, raw)
	}
	if strings.Contains(string(raw), "next_action") || len(env.Sources) != 0 {
		t.Fatal("explained omissions asked for recovery or audit sources", string(raw))
	}
}

func TestChangesRejectExplainedOmissionsWithDetailOrUnknownStatus(t *testing.T) {
	for name, change := range map[string]string{
		"title":        `{"resource_kind":"component","source_id":"gone","part":"editor","status":"deleted","title":"old"}`,
		"team":         `{"resource_kind":"component","source_id":"gone","part":"editor","status":"unselected","team_id":"team-1"}`,
		"reasons":      `{"resource_kind":"component","source_id":"gone","part":"editor","status":"deleted","reasons":[{"kind":"related","relation":null}]}`,
		"reason":       `{"resource_kind":"component","source_id":"gone","part":"editor","status":"deleted","reason":"not_in_current_capture"}`,
		"denied":       `{"resource_kind":"component","source_id":"gone","part":"editor","status":"denied"}`,
		"hidden":       `{"resource_kind":"component","source_id":"gone","part":"editor","status":"hidden"}`,
		"unknown kind": `{"resource_kind":"document","source_id":"gone","part":"editor","status":"deleted"}`,
	} {
		env := readChangesTest(t, changesStatesResult(`[`+change+`]`, true), "")
		if env.Error == nil || env.Data != nil {
			t.Fatalf("%s: unsafe omission accepted", name)
		}
	}
}
