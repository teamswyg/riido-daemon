package contextbroker

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

const changesFixture = `{"snapshot_id":"asg-1","since":"previous","changes":[],"complete":true,"full_refresh_required":false}`

func TestChangesUsesCurrentRunAndKeepsComparisonIdentity(t *testing.T) {
	b := New("", "")
	fetcher := &fakeFetcher{result: Result{Data: json.RawMessage(changesFixture), CapturedAt: time.Now()}}
	b.Fetcher = fetcher
	req := request(issueForTest(t, b, saasBinding()), "context changes")
	req.Args = map[string]string{"since": "previous"}
	env := b.Handle(req, activeSet("run-1"))
	if env.Error != nil {
		t.Fatal("comparison unavailable", env.Error.Code)
	}
	if len(fetcher.queries) != 1 || fetcher.queries[0].View != "changes" || fetcher.queries[0].AssignmentID != "asg-1" {
		t.Fatal("wrong current execution query")
	}
	raw, _ := json.Marshal(env.Data)
	var data map[string]any
	if json.Unmarshal(raw, &data) != nil || data["snapshot_id"] != "asg-1" || data["since"] != "previous" {
		t.Fatal("comparison identity lost")
	}
}

func TestChangesRefusalHasSnapshotRecovery(t *testing.T) {
	b := New("", "")
	b.Fetcher = &fakeFetcher{err: &FetchError{Status: 410, Message: "private upstream detail"}}
	req := request(issueForTest(t, b, saasBinding()), "context changes")
	req.Args = map[string]string{"since": "previous"}
	env := b.Handle(req, activeSet("run-1"))
	if env.Error == nil || env.Error.Code != contextcli.CodeSnapshotExpired || env.Error.Retryable || env.Error.NextAction != "riido context get" {
		t.Fatal("wrong baseline recovery", env.Error)
	}
}

func TestChangesInputIsBoundedBeforeFetch(t *testing.T) {
	for _, args := range []map[string]string{nil, {"since": "previous", "work": "other"}, {"since": "previous", "cursor": "other"}, {"since": "x\ny"}} {
		b := New("", "")
		fetcher := &fakeFetcher{}
		b.Fetcher = fetcher
		req := request(issueForTest(t, b, saasBinding()), "context changes")
		req.Args = args
		env := b.Handle(req, activeSet("run-1"))
		if env.Error == nil || env.Error.Code != contextcli.CodeUsage || len(fetcher.queries) != 0 {
			t.Fatal("invalid comparison reached server", env.Error)
		}
	}
}
