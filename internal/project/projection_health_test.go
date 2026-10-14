package project

import "testing"

func TestFromMwsdSnapshotReportsRepositoryHealth(t *testing.T) {
	snapshot := sampleSnapshot()
	snapshot.Projects.Repositories[1].RemoteMatches = false

	projection, err := FromMwsdSnapshot(snapshot)
	if err != nil {
		t.Fatalf("FromMwsdSnapshot returned error: %v", err)
	}
	if projection.Ready() {
		t.Fatal("projection should not be ready when a repo remote mismatches")
	}
	assertProjectHealth(t, projection, "gui_engine", RepositoryRemoteMismatch)
}

func assertProjectHealth(
	t *testing.T,
	projection WorkspaceProjection,
	projectID string,
	want RepositoryHealth,
) {
	t.Helper()
	for _, project := range projection.Projects {
		if project.ID == projectID {
			if project.Health != want {
				t.Fatalf("unexpected %s health: %s", projectID, project.Health)
			}
			return
		}
	}
	t.Fatalf("%s project missing", projectID)
}
