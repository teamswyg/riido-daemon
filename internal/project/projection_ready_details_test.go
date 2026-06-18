package project

import "testing"

func assertProjectionProjects(t *testing.T, projection WorkspaceProjection) {
	t.Helper()
	if len(projection.Projects) != 3 {
		t.Fatalf("unexpected project count: %d", len(projection.Projects))
	}
	if !projection.Ready() {
		t.Fatalf("projection should be ready, diagnostics=%v", projection.Diagnostics)
	}
	if projection.Projects[0].ID != "gui_engine" {
		t.Fatalf("projects should be sorted by id: %#v", projection.Projects)
	}
	for _, project := range projection.Projects {
		if project.Health != RepositoryReady {
			t.Fatalf("project %s should be ready, got %s", project.ID, project.Health)
		}
	}
}

func assertProjectionTaskLinks(t *testing.T, projection WorkspaceProjection) {
	t.Helper()
	if len(projection.DocumentTaskLinks) != 2 {
		t.Fatalf("unexpected document task link count: %d", len(projection.DocumentTaskLinks))
	}
	link := projection.DocumentTaskLinks[0]
	if link.TaskID != "task:mws.goal" ||
		link.ProjectID != "macmini-workspace" ||
		link.RecommendedProvider != "codex" ||
		link.RecommendedDecisionLLM != "codex" ||
		!link.RequiresHumanApproval {
		t.Fatalf("unexpected first task link: %+v", link)
	}
}
