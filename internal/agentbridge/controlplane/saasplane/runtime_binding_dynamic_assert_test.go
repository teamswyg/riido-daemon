package saasplane

import "testing"

func assertDynamicHeartbeat(t *testing.T, fake *fakeAssignmentServer, assignmentID string) {
	t.Helper()
	if len(fake.heartbeats) != 1 ||
		fake.heartbeats[0].RuntimeID != "daemon-1:codex" ||
		len(fake.heartbeats[0].ActiveAssignmentIDs) != 1 ||
		fake.heartbeats[0].ActiveAssignmentIDs[0] != assignmentID {
		t.Fatalf("dynamic heartbeats = %+v", fake.heartbeats)
	}
}
