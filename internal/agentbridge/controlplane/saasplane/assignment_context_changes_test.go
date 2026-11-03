package saasplane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchAssignmentContextOptsInToChangesStatesOnlyForChanges(t *testing.T) {
	seen := map[string][]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var view string
		if r.Header.Get("X-Riido-Context-Cache-Version") != "" {
			view = "content"
		}
		seen[view+"|"+r.Header.Get("X-Riido-Context-Changes-Version")] = r.Header.Values("X-Riido-Context-Changes-Version")
		_, _ = w.Write([]byte(`{"assignment_id":"asg-1","view":"changes","data":{},"sources":[]}`))
	}))
	defer server.Close()
	plane, err := New(Config{BaseURL: server.URL, DaemonID: "daemon-1", DeviceID: "device-1", DeviceSecret: "device-secret", RequestTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []AssignmentContextRequest{
		{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "changes", Since: "previous"},
		{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "content", Cursor: "c", CacheEnabled: true},
		{AssignmentID: "asg-1", LeaseToken: "lease-1", View: "task"},
	} {
		if _, err = plane.FetchAssignmentContext(context.Background(), "agent-1", in); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 3 || len(seen["|2"]) != 1 || seen["content|"] != nil || seen["|"] != nil {
		t.Fatalf("changes version header = %v", seen)
	}
}
