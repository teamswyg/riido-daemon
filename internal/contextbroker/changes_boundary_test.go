package contextbroker

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestChangesFailuresKeepCurrentRunBoundary(t *testing.T) {
	for _, test := range []struct {
		status  int
		message string
		code    contextcli.ErrorCode
	}{
		{400, `json: unknown field "since"`, contextcli.CodeControlPlaneUnsupported},
		{400, "invalid input", contextcli.CodeUsage},
		{403, "outside scope", contextcli.CodeOutOfRunScope},
		{403, "invalid lease", contextcli.CodePermissionDenied},
		{409, "terminal", contextcli.CodeLeaseLost},
		{429, "quota", contextcli.CodeRateLimited},
		{410, "expired", contextcli.CodeSnapshotExpired},
	} {
		b := New("", "")
		f := &fakeFetcher{err: &FetchError{Status: test.status, Message: test.message}}
		b.Fetcher = f
		req := request(issueForTest(t, b, saasBinding()), "context changes")
		req.Args = map[string]string{"since": "previous"}
		env := b.Handle(req, activeSet("run-1"))
		if env.Error == nil || env.Error.Code != test.code || env.Error.Retryable || len(f.queries) != 1 {
			t.Fatal("failure retried or remapped", test.status, env.Error)
		}
	}
}

func TestChangesEndedCapabilityNeverFetches(t *testing.T) {
	b := New("", "")
	f := &fakeFetcher{result: changedResult()}
	b.Fetcher = f
	req := request(issueForTest(t, b, saasBinding()), "context changes")
	req.Args = map[string]string{"since": "previous"}
	for range 2 {
		env := b.Handle(req, activeSet())
		if env.Error == nil || env.Error.Code != contextcli.CodePermissionDenied || len(f.queries) != 0 {
			t.Fatal("ended capability reached server")
		}
	}
}

func TestChangesRejectsDuplicatePartsAndOldMetadata(t *testing.T) {
	var data map[string]json.RawMessage
	if json.Unmarshal([]byte(changedFixture), &data) != nil {
		t.Fatal("fixture")
	}
	parts := strings.TrimSuffix(strings.TrimPrefix(string(data["changes"]), "["), "]")
	for _, value := range []string{"[" + parts + "," + parts + "]", `[{"resource_kind":"component","source_id":"old","part":"editor","status":"unavailable","reason":"not_in_current_capture","title":"OLD-HIDDEN"}]`} {
		r := changedResult()
		r.Sources = nil
		data["changes"] = json.RawMessage(value)
		data["complete"] = json.RawMessage("false")
		data["full_refresh_required"] = json.RawMessage("true")
		r.Data, _ = json.Marshal(data)
		env := readChangesTest(t, r, "")
		if env.Error == nil || env.Data != nil {
			t.Fatal("ambiguous or old metadata accepted")
		}
	}
}
