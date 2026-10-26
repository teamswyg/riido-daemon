package contextbroker

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestMeetingReadsUseSnapshotAndMeetingTarget(t *testing.T) {
	for _, tc := range []struct {
		command, view, target string
		args                  map[string]string
		body                  string
	}{
		{"meeting list", "meetings", "", nil, `{"meetings":[{"id":"m1","body_status":"available"}],"total":1,"truncated":false}`},
		{"meeting list", "meetings", "RIID-5900", map[string]string{"work": "RIID-5900"}, `{"meetings":[],"total":0,"truncated":false}`},
		{"meeting get", "meeting", "m1", map[string]string{"meeting": "m1"}, `{"meeting":{"id":"m1"},"document":{"content":"captured body","status":"available"}}`},
		{"meeting get", "meeting", "m2", map[string]string{"meeting": "m2"}, `{"meeting":{"id":"m2"},"document":{"content":"","status":"empty"}}`},
	} {
		t.Run(tc.command+"/"+tc.target, func(t *testing.T) {
			part := "body"
			if tc.view == "meetings" {
				part = "title"
			}
			fetcher := &fakeFetcher{result: Result{Data: json.RawMessage(tc.body), Sources: []ResultSource{{Kind: "related_meeting", ID: tc.target, Part: part, Version: "snapshot-v1"}}}}
			b := New("", "")
			b.Fetcher = fetcher
			req := request(issueForTest(t, b, saasBinding()), tc.command)
			req.Args = tc.args
			env := b.Handle(req, activeSet("run-1"))
			if env.Error != nil || len(env.Sources) != 1 || env.Sources[0].Version != "snapshot-v1" || env.Sources[0].Status != contextcli.SourceAvailable {
				t.Fatalf("meeting envelope = %+v", env)
			}
			if len(env.Warnings) == 0 || env.Warnings[0].Code != "ASSIGNMENT_SNAPSHOT" || env.Page != nil {
				t.Fatalf("meeting snapshot must not claim live/paged data: %+v", env)
			}
			encoded, _ := json.Marshal(env.Data)
			var got, want any
			if json.Unmarshal(encoded, &got) != nil || json.Unmarshal([]byte(tc.body), &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("meeting body or authoritative empty status was changed")
			}
			if q := fetcher.queries[0]; q.View != tc.view || q.Target != tc.target || q.AssignmentID != "asg-1" || q.LeaseToken != "lease-secret" {
				t.Fatalf("meeting request lost scoped routing: %+v", q)
			}
		})
	}
}
