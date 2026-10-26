package contextbroker

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestMeetingBodyBudgetAndPartialPreserveSource(t *testing.T) {
	for _, budget := range []string{"", "180"} {
		body, _ := json.Marshal(map[string]any{"meeting": map[string]string{"id": "m1"}, "document": map[string]string{"content": strings.Repeat("body", 250), "status": "partial"}})
		fetcher := &fakeFetcher{result: Result{Data: body, Truncated: budget == "", Sources: []ResultSource{{Kind: "related_meeting", ID: "m1", Part: "body", Version: "body-v1"}}}}
		b := New("", "")
		b.Fetcher = fetcher
		req := request(issueForTest(t, b, saasBinding()), "meeting get")
		req.Args = map[string]string{"meeting": "m1", "budget_chars": budget}
		env := b.Handle(req, activeSet("run-1"))
		if env.Error != nil || len(env.Sources) != 1 || env.Sources[0].Status != contextcli.SourcePartial || !env.Sources[0].Truncated || env.Sources[0].Version != "body-v1" {
			t.Fatalf("partial meeting source lost: %+v", env)
		}
		encoded, _ := json.Marshal(env.Data)
		if budget != "" && len(encoded) > 180 {
			t.Fatalf("meeting body exceeds budget: %d", len(encoded))
		}
		want := "SNAPSHOT_TRUNCATED"
		if budget != "" {
			want = "BUDGET_TRUNCATED"
		}
		found := false
		for _, warning := range env.Warnings {
			found = found || warning.Code == want
		}
		if !found {
			t.Fatalf("meeting partial warning %s missing", want)
		}
	}
}

func TestMeetingBodyFailuresRemainExplicit(t *testing.T) {
	for _, tc := range []struct {
		status  int
		message string
		code    contextcli.ErrorCode
	}{
		{403, "meeting is outside this assignment's context", contextcli.CodeOutOfRunScope},
		{403, "forbidden", contextcli.CodePermissionDenied},
		{404, "meeting body unavailable", contextcli.CodeNotFound},
		{503, "meeting body unavailable", contextcli.CodeUnavailable},
		{409, "assignment is not active", contextcli.CodeLeaseLost},
	} {
		b := New("", "")
		b.Fetcher = &fakeFetcher{err: &FetchError{Status: tc.status, Message: tc.message}}
		req := request(issueForTest(t, b, saasBinding()), "meeting get")
		req.Args = map[string]string{"meeting": "m1"}
		env := b.Handle(req, activeSet("run-1"))
		if env.Error == nil || env.Error.Code != tc.code || env.Data != nil || len(env.Sources) != 0 {
			t.Fatalf("failure became an empty or readable document: %+v", env)
		}
	}
}
