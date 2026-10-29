package contextbroker

import (
	"encoding/json"
	"testing"
)

func TestContentReadPreservesCapturedPageAndScope(t *testing.T) {
	fetcher := &fakeFetcher{result: Result{Data: json.RawMessage(`{"source_id":"work-one","part":"document","content":"captured page","next_cursor":"page-two"}`)}}
	b := New("", "")
	b.Fetcher = fetcher
	req := request(issueForTest(t, b, saasBinding()), "context part")
	req.Args = map[string]string{"work": "work-one", "cursor": "page-one"}
	env := b.Handle(req, activeSet("run-1"))
	if env.Error != nil || len(fetcher.queries) != 1 {
		t.Fatalf("captured page unavailable: %+v", env.Error)
	}
	q := fetcher.queries[0]
	if q.View != "content" || q.Target != "work-one" || q.Cursor != "page-one" || q.AssignmentID != "asg-1" || q.LeaseToken != "lease-secret" {
		t.Fatalf("wrong captured page scope: %+v", q)
	}
	if env.Page == nil || env.Page.NextCursor != "page-two" || env.Page.Limit != 1 {
		t.Fatalf("lost content continuation: %+v", env.Page)
	}
}
