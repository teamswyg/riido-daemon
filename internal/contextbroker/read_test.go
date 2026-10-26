package contextbroker

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type fakeFetcher struct {
	result  Result
	err     error
	queries []Query
}

func (f *fakeFetcher) FetchAssignmentContext(_ context.Context, q Query) (Result, error) {
	f.queries = append(f.queries, q)
	return f.result, f.err
}

func saasBinding() Binding {
	return Binding{RunKey: "run-1", TaskID: "task-1", WorkspaceID: "ws-1", AgentID: "agent-1", AssignmentID: "asg-1", LeaseToken: "lease-secret"}
}

func TestReadServesAssignmentSnapshot(t *testing.T) {
	fetcher := &fakeFetcher{result: Result{
		Data:       json.RawMessage(`{"component":{"keyNumber":"RIID-5900"}}`),
		CapturedAt: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC),
		Sources:    []ResultSource{{Kind: "task", ID: "task-1", Part: "body", Version: "sha256:abc"}},
	}}
	b := New("", "")
	b.Fetcher = fetcher
	token := issueForTest(t, b, saasBinding())
	req := request(token, "work get")
	req.Args = map[string]string{"work": "RIID-5900"}
	env := b.Handle(req, activeSet("run-1"))
	if env.Error != nil || len(env.Sources) != 1 || env.Sources[0].Version != "sha256:abc" || env.Sources[0].Status != contextcli.SourceAvailable {
		t.Fatalf("envelope=%+v", env)
	}
	if q := fetcher.queries[0]; q.View != viewTask || q.Target != "RIID-5900" || q.LeaseToken != "lease-secret" || q.AssignmentID != "asg-1" || q.RequestID != "cli_test" {
		t.Fatalf("query=%+v", q)
	}
	if encoded, _ := json.Marshal(b.Handle(request(token, "whoami"), activeSet("run-1"))); strings.Contains(string(encoded), "lease-secret") {
		t.Fatal("lease token must never leave the daemon")
	}
}

func TestReadMapsControlPlaneRefusals(t *testing.T) {
	cases := []struct {
		err  error
		want contextcli.ErrorCode
	}{
		{&FetchError{Status: 503, Message: "assignment context is not configured"}, contextcli.CodeControlPlaneUnsupported},
		{&FetchError{Status: 404, Message: "not found"}, contextcli.CodeControlPlaneUnsupported},
		{&FetchError{Status: 404, Message: "assignment context snapshot unavailable"}, contextcli.CodeNotFound},
		{&FetchError{Status: 403, Message: "target is outside this assignment's context"}, contextcli.CodeOutOfRunScope},
		{&FetchError{Status: 403, Message: "forbidden"}, contextcli.CodePermissionDenied},
		{&FetchError{Status: 409, Message: "assignment is not active"}, contextcli.CodeLeaseLost},
		{errors.New("dial tcp: refused"), contextcli.CodeUnavailable},
	}
	for _, tc := range cases {
		b := New("", "")
		b.Fetcher = &fakeFetcher{err: tc.err}
		token := issueForTest(t, b, saasBinding())
		if got := errorCode(t, b.Handle(request(token, "context get"), activeSet("run-1"))); got != tc.want {
			t.Fatalf("%v: got %s want %s", tc.err, got, tc.want)
		}
	}
}

func TestReadWithoutSnapshotSupportStaysUnsupported(t *testing.T) {
	b := New("", "")
	b.Fetcher = &fakeFetcher{}
	token := issueForTest(t, b, saasBinding())
	if got := errorCode(t, b.Handle(request(token, "work activity"), activeSet("run-1"))); got != contextcli.CodeControlPlaneUnsupported {
		t.Fatalf("activity is outside the snapshot, got %s", got)
	}
	local := New("", "")
	local.Fetcher = &fakeFetcher{}
	localToken := issueForTest(t, local, Binding{RunKey: "run-2"})
	if got := errorCode(t, local.Handle(request(localToken, "work get"), activeSet("run-2"))); got != contextcli.CodeControlPlaneUnsupported {
		t.Fatalf("non-SaaS runs have no lease, got %s", got)
	}
}

func TestReadCommentsPagesThroughCursor(t *testing.T) {
	fetcher := &fakeFetcher{result: Result{Data: json.RawMessage(`{"comments":[{"id":"c1"}],"total":25,"cursor":"0","next_cursor":"1"}`), Sources: []ResultSource{{Kind: "comment", ID: "c1", Part: "body"}}}}
	b := New("", "")
	b.Fetcher = fetcher
	token := issueForTest(t, b, saasBinding())
	req := request(token, "work comments")
	req.Args = map[string]string{"work": "RIID-5900", "cursor": "0"}
	env := b.Handle(req, activeSet("run-1"))
	if env.Error != nil || env.Page == nil || env.Page.NextCursor != "1" || !env.Page.Truncated {
		t.Fatalf("envelope=%+v page=%+v", env, env.Page)
	}
	if q := fetcher.queries[0]; q.View != viewComments || q.Cursor != "0" {
		t.Fatalf("query=%+v", q)
	}
}

func TestReadAgentsDirectoryIsLiveAndPaged(t *testing.T) {
	fetcher := &fakeFetcher{result: Result{
		Live:       true,
		CapturedAt: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC),
		Data:       json.RawMessage(`{"agents":[{"agent_id":"agent-b"}],"total":60,"cursor":"0","next_cursor":"1"}`),
		Sources:    []ResultSource{{Kind: "agent", ID: "agent-b", Part: "profile"}},
	}}
	b := New("", "")
	b.Fetcher = fetcher
	token := issueForTest(t, b, saasBinding())
	env := b.Handle(request(token, "agents list"), activeSet("run-1"))
	if env.Error != nil || env.Page == nil || env.Page.Limit != agentPageSize || env.Page.NextCursor != "1" {
		t.Fatalf("envelope=%+v page=%+v", env, env.Page)
	}
	if len(env.Warnings) == 0 || env.Warnings[0].Code != "LIVE_STATUS" {
		t.Fatalf("live answer must not claim snapshot freshness: %+v", env.Warnings)
	}
	req := request(token, "agents show")
	req.Args = map[string]string{"agent": "agent-b"}
	b.Handle(req, activeSet("run-1"))
	if q := fetcher.queries[len(fetcher.queries)-1]; q.View != viewAgent || q.Target != "agent-b" {
		t.Fatalf("query=%+v", q)
	}
}
